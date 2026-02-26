package services

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"database/sql"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"slices"
	db_image "social-network/pkg/db/queries/image"
	db_public_images "social-network/pkg/db/queries/public_images"
	"social-network/pkg/db/sqlite"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
)

var secretKey = []byte(os.Getenv("SECRET_SIGN"))

type FileService struct {
	BasePath   string
	DB         *sql.DB
	Interval   time.Duration
	stopChan   chan bool
	Logger     *slog.Logger
	PublicPath string
}

type File struct {
	UUId      string
	User      []byte
	Filename  string
	ImagePath string
	CreatedAt time.Time
	ExpiresAt time.Time
}

type PublicFile struct {
	UUId         string
	GuestSession []byte
	Filename     string
	ImagePath    string
	CreatedAt    time.Time
	ExpiresAt    time.Time
}

type SaveDBImage interface {
	saveImageToDatabase(db *sql.DB) error
}

func NewFileService(privatePath string, publicPath string, db *sql.DB, interval time.Duration, log *slog.Logger) *FileService {
	return &FileService{
		BasePath:   privatePath,
		DB:         db,
		Interval:   interval,
		stopChan:   make(chan bool),
		Logger:     log,
		PublicPath: publicPath,
	}
}

func (s *FileService) StartCleanUp() {

	ticker := time.NewTicker(s.Interval)
	s.Logger.Info("started clean up service")

	go func() {
		s.runCleanUp()

		for {
			select {
			case <-ticker.C:
				s.runCleanUp()
			case <-s.stopChan:
				ticker.Stop()
				return
			}
		}
	}()

}

func (s *FileService) StopCleanUp() {
	s.stopChan <- true
}

func (s *FileService) runCleanUp() {
	ctx := context.Background()
	start := time.Now().UTC()
	s.Logger.Info("started clean up images")

	images, err := sqlite.NewQuery(s.DB).Image.GetNotSetImages(ctx)

	if err != nil {
		s.Logger.Error("error fetching images")
		return
	}

	publicImages, err := sqlite.NewQuery(s.DB).PublicImage.GetPublicImages(ctx)

	if err != nil {
		s.Logger.Error("error fetching public images")
		return
	}

	imagesToDelete := []string{}
	imagePaths := []string{}

	for _, image := range images {
		if !start.After(image.ExpiresAt.Time) {
			continue
		}

		imagesToDelete = append(imagesToDelete, image.ImageID)
		imagePaths = append(imagePaths, image.ImagePath)
	}

	for _, publicImage := range publicImages {
		if !start.After(publicImage.ExpiresAt.Time) {
			continue
		}

		imagesToDelete = append(imagesToDelete, publicImage.ImageID)
		imagePaths = append(imagePaths, publicImage.ImagePath)
	}

	if len(imagesToDelete) == 0 {
		return
	}

	tx, err := s.DB.BeginTx(ctx, nil)
	if err != nil {
		s.Logger.Error("failed to begin transaction", "error", err)
		return
	}

	err = sqlite.NewQuery(s.DB).Image.WithTx(tx).DeleteImages(ctx, imagesToDelete)

	if err != nil {
		tx.Rollback()
		s.Logger.Error("failed to delete images from database", "error", err)
		return
	}

	err = sqlite.NewQuery(s.DB).PublicImage.WithTx(tx).DeletePublicImages(ctx, imagesToDelete)

	if err != nil {
		tx.Rollback()
		s.Logger.Error("failed to delete images from database", "error", err)
		return
	}

	if err := tx.Commit(); err != nil {
		s.Logger.Error("failed to commit transaction", "error", err)
		return
	}

	s.Logger.Info("deleted images from database", "count", len(imagesToDelete))

	deletedFiles := 0
	for _, imagePath := range imagePaths {
		if err := os.Remove(imagePath); err != nil {
			s.Logger.Warn("failed to delete file (will be recovered on next startup)",
				"path", imagePath, "error", err)
		} else {
			deletedFiles++
		}
	}

	s.Logger.Info("cleanup completed", "database_records", len(imagesToDelete), "files deleted", deletedFiles)
}

func (s *FileService) UploadHandler(file multipart.File, user []byte, path string) (SaveDBImage, error) {

	if path != s.BasePath && path != s.PublicPath {
		return nil, errors.New("wrong path")
	}

	if _, err := os.Stat(path); os.IsNotExist(err) {
		os.Mkdir(path, 0755)
	}

	readFile, err := io.ReadAll(file)

	if err != nil {
		return nil, err
	}

	extension, err := s.fileExtension(readFile)

	if err != nil {
		return nil, err
	}

	uuid := uuid.New().String()

	filename := uuid + "." + extension

	ok := s.isValidFile(readFile)

	if !ok {
		return nil, errors.New("not an image")
	}

	create, err := os.Create(filepath.Join(path, filename))

	if err != nil {
		return nil, err
	}

	_, err = create.Write(readFile)

	if err != nil {
		return nil, err
	}

	IsPublic := strings.Contains(path, "public")

	var img SaveDBImage

	if IsPublic {
		img = &PublicFile{
			UUId:         uuid,
			GuestSession: user,
			Filename:     filename,
			ImagePath:    filepath.Join(path, filename),
			CreatedAt:    time.Now().UTC(),
			ExpiresAt:    time.Now().UTC().Add(5 * time.Minute),
		}
	} else {
		img = &File{
			UUId:      uuid,
			User:      user,
			Filename:  filename,
			ImagePath: filepath.Join(path, filename),
			CreatedAt: time.Now().UTC(),
			ExpiresAt: time.Now().UTC().Add(5 * time.Minute),
		}
	}

	err = s.saveImage(img)

	if err != nil {
		return nil, err
	}

	return img, nil

}

func (s *FileService) AssignImage(imageUUId string) error {
	err := sqlite.NewQuery(s.DB).Image.AssignImage(context.Background(), imageUUId)

	if err != nil {
		return err
	}

	return nil
}

func (s *FileService) RemoveImage(imageUUId string) error {
	err := sqlite.NewQuery(s.DB).Image.SetImageExpiry(context.Background(), db_image.SetImageExpiryParams{
		Column1: time.Now().UTC().Format("2006-01-02 15:04:05"),
		ImageID: imageUUId,
	})

	if err != nil {
		return err
	}

	return nil
}

func (s *FileService) GenerateSignImage(filename string, userID []byte, expires time.Time) string {

	signMessage := fmt.Sprintf("%s:%d", userID, expires.Unix())

	mac := hmac.New(sha256.New, secretKey)

	mac.Write([]byte(signMessage))

	signature := base64.URLEncoding.EncodeToString(mac.Sum(nil))

	url := os.Getenv("APP_URL")

	return fmt.Sprintf("%s/api/storage/image/%s?expires=%d&signature=%s", url, filename, expires.Unix(), signature)
}

func (s *FileService) GeneratePublicSignImage(filename string, userID []byte, expires time.Time) string {

	signMessage := fmt.Sprintf("%s:%d", userID, expires.Unix())

	mac := hmac.New(sha256.New, secretKey)

	mac.Write([]byte(signMessage))

	signature := base64.URLEncoding.EncodeToString(mac.Sum(nil))

	url := os.Getenv("APP_URL")

	return fmt.Sprintf("%s/api/public/image/%s?expires=%d&signature=%s", url, filename, expires.Unix(), signature)
}

func (s *FileService) ValidateImageSign(r *http.Request, userId []byte) (bool, error) {

	queries := r.URL.Query()
	expires := queries.Get("expires")
	signature := queries.Get("signature")

	if expires == "" && signature == "" {
		return false, errors.New("missing paramters")
	}

	convert, err := strconv.ParseInt(expires, 10, 64)

	if err != nil {
		return false, errors.New("failed to convert duration")
	}

	if time.Now().Unix() > convert {
		return false, errors.New("expired")
	}

	message := fmt.Sprintf("%s:%d", userId, convert)
	mac := hmac.New(sha256.New, secretKey)
	mac.Write([]byte(message))
	expectedSig := base64.URLEncoding.EncodeToString(mac.Sum(nil))

	return hmac.Equal([]byte(signature), []byte(expectedSig)), nil
}

func (s *FileService) MoveToPrivate(img string, user []byte, extension string) error {
	ctx := context.Background()

	tx, err := s.DB.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if _, err := os.Stat(s.BasePath); os.IsNotExist(err) {
		os.Mkdir(s.BasePath, 0755)
	}

	if _, err := os.Stat(s.PublicPath); os.IsNotExist(err) {
		os.Mkdir(s.BasePath, 0755)
	}

	getPublicImage, err := sqlite.NewQuery(s.DB).PublicImage.GetPublicImage(ctx, img)
	if err != nil {
		return err
	}

	publicFolderFile, err := os.Open(filepath.Join(s.PublicPath, img+"."+extension))
	if err != nil {
		return err
	}
	defer publicFolderFile.Close()

	privateFolderFile, err := os.Create(filepath.Join(s.BasePath, img+"."+extension))
	if err != nil {
		return err
	}
	defer privateFolderFile.Close()

	_, err = io.Copy(privateFolderFile, publicFolderFile)
	if err != nil {
		os.Remove(filepath.Join(s.BasePath, img))
		return err
	}

	savePrivateFile := &File{
		UUId:      getPublicImage.ImageID,
		User:      user,
		Filename:  getPublicImage.FileName,
		ImagePath: filepath.Join(s.BasePath, img),
		CreatedAt: getPublicImage.CreatedAt.Time,
		ExpiresAt: time.Now().UTC().Add(24 * time.Hour),
	}

	err = sqlite.NewQuery(s.DB).Image.WithTx(tx).CreateImage(ctx, db_image.CreateImageParams{
		ImageID:   savePrivateFile.UUId,
		PosterID:  savePrivateFile.User,
		ImagePath: savePrivateFile.ImagePath,
		FileName:  savePrivateFile.Filename,
		Column5:   savePrivateFile.CreatedAt.Format("2006-01-02 15:04:05"),
		Column6:   savePrivateFile.ExpiresAt.Format("2006-01-02 15:04:05"),
	})

	if err != nil {
		os.Remove(filepath.Join(s.BasePath, img+"."+extension))
		return err
	}

	err = sqlite.NewQuery(s.DB).Image.WithTx(tx).AssignImage(ctx, savePrivateFile.UUId)
	if err != nil {
		os.Remove(filepath.Join(s.BasePath, img+"."+extension))
		return err
	}

	err = sqlite.NewQuery(s.DB).PublicImage.WithTx(tx).DeletePublicImage(ctx, getPublicImage.ImageID)
	if err != nil {
		os.Remove(filepath.Join(s.BasePath, img+"."+extension))
		return err
	}

	if err := tx.Commit(); err != nil {
		os.Remove(filepath.Join(s.BasePath, img+"."+extension))
		return err
	}

	err = os.Remove(filepath.Join(s.PublicPath, img+"."+extension))
	if err != nil {
		s.Logger.Warn("failed to delete public file", "path", filepath.Join(s.PublicPath, img))
	}

	return nil
}

func (s *FileService) saveImage(img SaveDBImage) error {
	return img.saveImageToDatabase(s.DB)
}

func (file *File) saveImageToDatabase(db *sql.DB) error {
	err := sqlite.NewQuery(db).Image.CreateImage(context.Background(), db_image.CreateImageParams{
		ImageID:   file.UUId,
		PosterID:  file.User,
		ImagePath: file.ImagePath,
		FileName:  file.Filename,
		Column5:   file.CreatedAt.Format("2006-01-02 15:04:05"),
		Column6:   file.ExpiresAt.Format("2006-01-02 15:04:05"),
	})

	if err != nil {
		return err
	}

	return nil
}

func (publicFile *PublicFile) saveImageToDatabase(db *sql.DB) error {
	err := sqlite.NewQuery(db).PublicImage.SavePublicImage(context.Background(), db_public_images.SavePublicImageParams{
		ImageID:      publicFile.UUId,
		GuestSession: publicFile.GuestSession,
		ImagePath:    publicFile.ImagePath,
		FileName:     publicFile.Filename,
		Column5:      publicFile.CreatedAt.Format("2006-01-02 15:04:05"),
		Column6:      publicFile.ExpiresAt.Format("2006-01-02 15:04:05"),
	})

	if err != nil {
		return err
	}

	return nil

}

func (s *FileService) isValidFile(file []byte) bool {
	filetype := http.DetectContentType(file)

	return strings.HasPrefix(filetype, "image/")
}

func (s *FileService) fileExtension(file []byte) (string, error) {
	filetype := http.DetectContentType(file)

	allowedFiles := []string{"png", "jpg", "gif", "jpeg"}

	extension := strings.Split(filetype, "/")

	if len(extension) != 2 {
		return "", errors.New("wrong file format")
	}

	if !slices.Contains(allowedFiles, extension[1]) {
		return "", errors.New("wrong file extension")
	}

	return extension[1], nil
}
