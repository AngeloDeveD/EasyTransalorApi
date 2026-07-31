package files

import (
	"errors"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/google/uuid"
)

type FileRepository interface {
	SaveArchive(gameid int, file *multipart.FileHeader) (string, error)
	SaveImages(big_image *multipart.FileHeader, small_image *multipart.FileHeader) ([]string, error)
	IsAllowedArchiveFormat(file *multipart.FileHeader) error
	IsAllowedImageFormat(file *multipart.FileHeader) error
	IsAllowedArchiveSize(fileSize int64) error
	IsAllowedImageSize(small_image_size int64, big_image_size int64) error
	DeleteArchiveFile(gameId int) error
	DeleteImageFiles(big_img_url string) error
	DeleteOneImageFile(fileUrl string) error
	GetArchivePath(titleFile string, author string, fileUrl string) (string, string, error)
}

// Структруа files
type LocalFileRepo struct{}

func NewLocalFileRepo() *LocalFileRepo {
	return &LocalFileRepo{}
}

var gb5 int64 = 5 << 30 //5гб
var mb5 int64 = 5 << 20 //5мб

func (r *LocalFileRepo) IsAllowedArchiveFormat(file *multipart.FileHeader) error {
	allowFileExts := map[string]bool{".zip": true, ".7zip": true, ".rar": true}
	ext := strings.ToLower(filepath.Ext(file.Filename))
	if !allowFileExts[ext] {
		return errors.New("неподдерживаемый формат файла!")
	}
	return nil
}

func (r *LocalFileRepo) IsAllowedImageFormat(image *multipart.FileHeader) error {
	allowImagesExts := map[string]bool{".jpg": true, ".jpeg": true, ".png": true, ".webp": true}
	ext := strings.ToLower(filepath.Ext(image.Filename))
	if !allowImagesExts[ext] {
		return errors.New("неподдерживаемый формат изображения")
	}
	return nil
}

func (r *LocalFileRepo) IsAllowedArchiveSize(fileSize int64) error {
	if fileSize > gb5 {
		return errors.New("файл слишком большой!")
	}
	return nil
}

func (r *LocalFileRepo) IsAllowedImageSize(small_image_size int64, big_image_size int64) error {
	if small_image_size > mb5 || big_image_size > mb5 {
		return errors.New("размер изображения большой")
	}
	return nil
}

// Вспомогательная функция для сохранения файла без Gin
func saveFile(file *multipart.FileHeader, savePath string) error {
	// 1. Открываем входящий поток файла
	src, err := file.Open()
	if err != nil {
		return err
	}
	defer src.Close()

	// 2. Создаем файл на диске
	dst, err := os.Create(savePath)
	if err != nil {
		return err
	}
	defer dst.Close()

	// 3. Копируем байты из входящего потока в файл на диске
	if _, err = io.Copy(dst, src); err != nil {
		return err
	}
	return nil
}

func (r *LocalFileRepo) SaveArchive(gameid int, file *multipart.FileHeader) (string, error) {
	ext := strings.ToLower(filepath.Ext(file.Filename))
	newFile := uuid.New().String() + ext

	folderPath := filepath.Join("uploads", "files", strconv.Itoa(gameid))

	if err := os.MkdirAll(folderPath, os.ModePerm); err != nil {
		return "", errors.New("Не удалось создать директорию!")
	}

	savePathFile := filepath.Join(folderPath, newFile)

	if err := saveFile(file, savePathFile); err != nil {
		return "", errors.New("не удалось сохранить файл")
	}

	savePathFile = strings.ReplaceAll(savePathFile, `\`, "/")
	file_url := strings.Replace(savePathFile, "uploads", "static", 1)

	return file_url, nil
}

func (r *LocalFileRepo) SaveImages(big_image *multipart.FileHeader, small_image *multipart.FileHeader) ([]string, error) {
	if big_image.Size > mb5 || small_image.Size > mb5 {
		return []string{}, errors.New("файл слишком большой!")
	}

	ext_big := strings.ToLower(filepath.Ext(big_image.Filename))
	ext_small := strings.ToLower(filepath.Ext(small_image.Filename))

	newPic := uuid.New().String()
	newPic_big := newPic + ext_big
	newPic_small := newPic + ext_small

	bigDir := filepath.Join("uploads", "Icons", "Big")
	smallDir := filepath.Join("uploads", "Icons", "Small")

	// Создаём каталоги, если их ещё нет (как в SaveArchive) — иначе os.Create упадёт.
	if err := os.MkdirAll(bigDir, os.ModePerm); err != nil {
		return []string{}, errors.New("не удалось создать директорию для иконок")
	}
	if err := os.MkdirAll(smallDir, os.ModePerm); err != nil {
		return []string{}, errors.New("не удалось создать директорию для иконок")
	}

	savePathBig := filepath.Join(bigDir, newPic_big)
	savePathSmall := filepath.Join(smallDir, newPic_small)

	if err := saveFile(big_image, savePathBig); err != nil {
		return []string{}, errors.New("Ошибка сохранения большой иконки")
	}

	if err := saveFile(small_image, savePathSmall); err != nil {
		return []string{}, errors.New("Ошибка сохранения маленькой иконки")
	}

	savePathSmall = strings.ReplaceAll(savePathSmall, `\`, "/")
	savePathBig = strings.ReplaceAll(savePathBig, `\`, "/")

	image_small_url := strings.Replace(savePathSmall, "uploads", "static", 1)
	image_big_url := strings.Replace(savePathBig, "uploads", "static", 1)

	return []string{image_small_url, image_big_url}, nil
}

func (r *LocalFileRepo) DeleteArchiveFile(gameId int) error {
	gameId_str := strconv.Itoa(gameId)
	folderPath := filepath.Join("uploads", "files", gameId_str)
	if err := os.RemoveAll(folderPath); err != nil {
		return errors.New("ошибка удаления архива с файлами")
	}
	return nil
}

func (r *LocalFileRepo) DeleteOneImageFile(fileUrl string) error {

	fileUrl = strings.ReplaceAll(fileUrl, `\`, "/")

	filePath := strings.Replace(fileUrl, "static", "uploads", 1)
	if err := os.Remove(filePath); err != nil {
		return errors.New("Ошибка удаления изображения")
	}

	return nil
}

func (r *LocalFileRepo) DeleteImageFiles(big_img_url string) error {
	filePathBig := strings.Replace(big_img_url, "/static/", "uploads/", 1)
	filePathSmall := strings.Replace(filePathBig, "/Big/", "/Small/", 1)

	if err := os.Remove(filePathBig); err != nil {
		return errors.New("ошибка удаления большого изображения")
	}

	if err := os.Remove(filePathSmall); err != nil {
		return errors.New("ошибка удаления маленького изображения")
	}

	return nil
}

func (r *LocalFileRepo) GetArchivePath(titleFile string, author string, fileUrl string) (string, string, error) {
	if fileUrl == "" {
		return "", "", errors.New("URL файла пуст!")
	}

	// 1. Превращаем URL в реальный путь на диске
	filePath := strings.Replace(fileUrl, "/static/", "uploads/", 1)

	// 2. Формируем красивое имя для скачивания
	rawName := titleFile + "_" + author + filepath.Ext(fileUrl)
	filename := strings.ReplaceAll(rawName, " ", "_")
	filename = strings.ReplaceAll(filename, "+", "_")

	// Возвращаем путь на диске и красивое имя
	return filePath, filename, nil
}
