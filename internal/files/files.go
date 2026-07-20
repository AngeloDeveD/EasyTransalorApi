package files

import (
	"errors"
	"mime/multipart"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type FileRepository interface {
	//Сохранение файлов
	SaveArchive(gameid int, file *multipart.FileHeader, c *gin.Context) (string, error)
	SaveImages(title string) error
	//Проверка поддерживаемый форматов
	IsAllowedArchiveFormat(file *multipart.FileHeader) error
	//Проверка размера файлов
	IsAllowedArchiveSize(fileSize int64) error
	//Удаление файла
}

/*Проверка файла на поддерживаемый формат файлов*/
func IsAllowedArchiveFormat(file *multipart.FileHeader) error {

	//Поддерживаемые форматы файлов
	allowFileExts := map[string]bool{
		".zip":  true,
		".7zip": true,
		".rar":  true,
	}

	//Получение текущего формата файла
	ext := strings.ToLower(filepath.Ext(file.Filename))

	if !allowFileExts[ext] {
		return errors.New("Неподдерживаемый формат файла!")
	}
	return nil
}

/*Проверка на допустимсы размер файла*/
func IsAllowedArchiveSize(fileSize int64) error {
	if fileSize > 5<<30 { //Если файл больше 5 гб
		return errors.New("Файл слишком большой!")
	}

	return nil
}

/*Сохраняет архив перевода в папку и возвращает пусть, в котором сохранён файл*/
func SaveArchive(gameid int, file *multipart.FileHeader, c *gin.Context) (string, error) {

	ext := strings.ToLower(filepath.Ext(file.Filename))
	newFile := uuid.New().String() + ext

	//путь: /uploads/files/(gameid)
	folderPath := filepath.Join("uploads", "files", strconv.Itoa(gameid))

	//Проверка на существование директории
	if err := os.MkdirAll(folderPath, os.ModePerm); err != nil {
		return "", errors.New("Не удалось создать директорию!")
	}

	savePathFile := filepath.Join(folderPath, newFile)

	//Сохранение файла
	if err := c.SaveUploadedFile(file, savePathFile); err != nil {
		return "", errors.New("Не удалось сохранить файл")
	}

	//Для windows
	savePathFile = strings.ReplaceAll(savePathFile, `\`, "/")

	return savePathFile, nil
}
