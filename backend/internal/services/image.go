package services

import (
	"context"
	"database/sql"
	"errors"
	"io"
	"log/slog"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"slices"
	db_image "social-network/pkg/db/queries/image"
	"social-network/pkg/db/sqlite"
	"strings"
	"time"

	"github.com/google/uuid"
)

type FileService struct {
	BasePath string
	DB       *sql.DB
	Interval time.Duration
	stopChan chan bool
	Logger   *slog.Logger
}

type File struct {
	UUId      string
	User      []byte
	ImagePath string
	CreatedAt time.Time
	ExpiresAt time.Time
}

func NewFileService(path string, db *sql.DB, interval time.Duration, log *slog.Logger) *FileService {
	return &FileService{
		BasePath: path,
		DB:       db,
		Interval: interval,
		stopChan: make(chan bool),
		Logger:   log,
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
	start := time.Now()
	s.Logger.Info("started clean up images")

	images, err := sqlite.NewQuery(s.DB).Image.GetNotSetImages(ctx)

	if err != nil {
		s.Logger.Error("error fetching images")
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

func (s *FileService) UploadHandler(file multipart.File, user []byte) (*File, error) {

	if _, err := os.Stat(s.BasePath); os.IsNotExist(err) {
		os.Mkdir(s.BasePath, 0755)
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

	create, err := os.Create(filepath.Join(s.BasePath, filename))

	if err != nil {
		return nil, err
	}

	_, err = create.Write(readFile)

	if err != nil {
		return nil, err
	}

	databaseFile := File{
		UUId:      uuid,
		User:      user,
		ImagePath: s.BasePath + "/" + filename,
		CreatedAt: time.Now(),
		ExpiresAt: time.Now().Add(5 * time.Minute),
	}

	err = s.saveFileToDatabase(databaseFile)

	if err != nil {
		return nil, err
	}

	return &databaseFile, nil

}

func (s *FileService) AssignImage(imageUUId string) error {
	err := sqlite.NewQuery(s.DB).Image.ImageState(context.Background(), db_image.ImageStateParams{
		ExpiresAt: sql.NullTime{},
		ImageID:   imageUUId,
	})

	if err != nil {
		return err
	}

	return nil
}

func (s *FileService) RemoveImage(imageUUId string) error {
	err := sqlite.NewQuery(s.DB).Image.ImageState(context.Background(), db_image.ImageStateParams{
		ExpiresAt: sql.NullTime{Time: time.Now(), Valid: true},
		ImageID:   imageUUId,
	})

	if err != nil {
		return err
	}

	return nil
}

func (s *FileService) saveFileToDatabase(file File) error {
	err := sqlite.NewQuery(s.DB).Image.CreateImage(context.Background(), db_image.CreateImageParams{
		ImageID:   file.UUId,
		PosterID:  file.User,
		ImagePath: file.ImagePath,
		CreatedAt: sql.NullTime{Time: file.CreatedAt, Valid: true},
		ExpiresAt: sql.NullTime{Time: file.ExpiresAt, Valid: true},
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

	alowedFiles := []string{"png", "jpg", "gif"}

	extension := strings.Split(filetype, "/")

	if len(extension) != 2 {
		return "", errors.New("wrong file")
	}

	if !slices.Contains(alowedFiles, extension[1]) {
		return "", errors.New("wrong file")
	}

	return extension[1], nil
}
