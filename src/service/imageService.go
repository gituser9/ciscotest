package service

import (
	"bytes"
	"encoding/base64"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"strings"

	"github.com/disintegration/imaging"
	"github.com/google/uuid"
	"github.com/kpango/glg"

	"model"
)

type ImageService struct {
	resizeChan chan model.ResizeData
	config     *model.Config
	isResizing bool // for test only
}

func CreateImageService(config *model.Config) *ImageService {
	service := new(ImageService)
	service.config = config
	service.resizeChan = make(chan model.ResizeData) // not buffered
	go service.resizeImage()

	return service
}

func (s *ImageService) DecodeFromBase64(base64String string, imageType string) bool {
	formats := getFormats()
	lowerImageType := strings.ToLower(imageType)
	format, ok := formats[lowerImageType]

	if !ok {
		glg.Errorf("%s : %s", "Unknown format", imageType)
		return false
	}

	imageBytes, err := base64.StdEncoding.DecodeString(base64String)

	if err != nil {
		glg.Errorf("%s : %s", "Base64 string decode", err.Error())
		return false
	}

	reader := bytes.NewReader(imageBytes)
	decoded, err := imaging.Decode(reader)

	if err != nil {
		glg.Errorf("%s : %s", "Decode", err.Error())
		return false
	}

	filePath := s.config.ImageDirectory + uuid.New().String() + "." + lowerImageType
	file, err := os.OpenFile(filePath, os.O_WRONLY|os.O_CREATE, 0777)

	if err != nil {
		glg.Errorf("%s : %s", "Create file", err.Error())
		return false
	}

	imaging.Encode(file, decoded, format)
	file.Close()

	s.resizeChan <- model.ResizeData{FilePath: filePath, Format: imageType}
	return true
}

func (s *ImageService) GetFromMultipartFormData(form *multipart.Form) (bool, []string) {
	files := form.File["images"]
	var errorNames []string

	for _, item := range files {
		file, err := item.Open()

		if err != nil {
			glg.Errorf("Open image", err.Error())
			errorNames = append(errorNames, item.Filename)
			continue
		}

		filePath := s.config.ImageDirectory + item.Filename
		dst, err := os.Create(filePath)

		if err != nil {
			glg.Errorf("Create file for save uploaded", err.Error())
			errorNames = append(errorNames, item.Filename)
			continue
		}
		if _, err := io.Copy(dst, file); err != nil {
			glg.Errorf("Copy", err.Error())
			errorNames = append(errorNames, item.Filename)
			continue
		}

		s.resizeChan <- model.ResizeData{FilePath: filePath, Format: getFormat(filePath)}

		dst.Close()
		file.Close()
	}

	return len(errorNames) == 0, errorNames
}

func (s *ImageService) GetByUrl(url string) bool {
	response, err := http.Get(url)

	if err != nil {
		glg.Errorf("%s : %s", "Get from URL", err.Error())
		return false
	}
	if response.StatusCode != http.StatusOK {
		glg.Errorf("%s : %s", "Bad request", err.Error())
		return false
	}

	filePath := s.config.ImageDirectory + getFileName(url)
	file, err := os.Create(filePath)

	if err != nil {
		glg.Errorf("%s : %s", "Create file", err.Error())
		return false
	}

	defer file.Close()
	_, err = io.Copy(file, response.Body)

	if err != nil {
		glg.Errorf("%s : %s", "Copy", err.Error())
		return false
	}

	s.resizeChan <- model.ResizeData{FilePath: filePath, Format: strings.ToLower(getFormat(url))}
	return true
}

func (s *ImageService) resizeImage() {
	for {
		select {
		case imageData := <-s.resizeChan:
			s.isResizing = true
			sourceImage, err := imaging.Open(imageData.FilePath)
			name := getFileName(imageData.FilePath)

			if err != nil {
				glg.Errorf("%s : %s", "Open saved image for resize", err.Error())
				s.isResizing = false
				break
			}

			format, ok := getImagingFormat(imageData.Format)

			if !ok {
				glg.Errorf("%s : %s", "Unknown format", imageData.Format)
				s.isResizing = false
				break
			}

			resizeNRGBA := imaging.Resize(sourceImage, s.config.ImageWidth, s.config.ImageHeight, imaging.Linear)
			file, err := os.OpenFile(s.config.ResizedImageDirectory+name, os.O_WRONLY|os.O_CREATE, 0777)

			if err != nil {
				glg.Errorf("%s : %s", "Create file for resized image", err.Error())
				s.isResizing = false
				break
			}

			imaging.Encode(file, resizeNRGBA, format)
			file.Close()
			s.isResizing = false
		}
	}
}

func getFileName(filePath string) string {
	pathChunks := strings.FieldsFunc(filePath, split)
	return pathChunks[len(pathChunks)-1]
}

func getFormat(filePath string) string {
	pathChunks := strings.Split(filePath, ".")
	return pathChunks[len(pathChunks)-1]
}

func split(item rune) bool {
	return item == '/' || item == '\''
}

func getFormats() map[string]imaging.Format {
	return map[string]imaging.Format{
		"jpeg": imaging.JPEG,
		"jpg":  imaging.JPEG,
		"gif":  imaging.GIF,
		"bmp":  imaging.BMP,
		"tiff": imaging.TIFF,
		"png":  imaging.PNG,
	}
}

func getImagingFormat(stringFormat string) (imaging.Format, bool) {
	formats := getFormats()
	format, ok := formats[strings.ToLower(stringFormat)]

	return format, ok
}
