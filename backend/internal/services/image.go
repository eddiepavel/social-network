package services

import (
	"errors"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/google/uuid"
)

type FileService struct {
	BasePath string
	File     multipart.File
}

func NewFileService(path string, file multipart.File) *FileService {
	return &FileService{
		BasePath: path,
		File:     file,
	}
}

func (s *FileService) Upload() (string, error) {

	if _, err := os.Stat(s.BasePath); os.IsNotExist(err) {
		os.Mkdir(s.BasePath, 0755)
	}

	file, err := io.ReadAll(s.File)

	if err != nil {
		return "", err
	}

	extension, err := s.fileExtension(file)

	if err != nil {
		return "", err
	}

	uuid := uuid.New().String()

	filename := uuid + "." + extension

	ok := s.isValidFile(file)

	if !ok {
		return "", errors.New("not an image")
	}

	create, err := os.Create(filepath.Join(s.BasePath, filename))

	if err != nil {
		return "", err
	}

	_, err = create.Write(file)

	if err != nil {
		return "", err
	}

	return filename, nil

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
