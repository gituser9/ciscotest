package server

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"strings"

	"model"
	"service"

	"github.com/kpango/glg"
)

const contentTypeHeader = "Content-Type"
const jsonContentType = "application/json"
const multipartFormDataContentType = "multipart/form-data"

type Server struct {
	config       *model.Config
	imageService *service.ImageService
}

func CreateServer(config *model.Config) *Server {
	server := new(Server)
	server.config = config
	server.imageService = service.CreateImageService(config)

	return server
}

func (s *Server) Start() {
	http.HandleFunc("/", s.handler)

	address := ":" + strconv.Itoa(s.config.Port)
	glg.Infof("Server is started on %d port", s.config.Port)

	if err := http.ListenAndServe(address, nil); err != nil {
		log.Fatal(err)
	}
}

func (s Server) handler(w http.ResponseWriter, r *http.Request) {
	w.Header().Add(contentTypeHeader, jsonContentType)

	if r.Method != http.MethodPost && r.Method != http.MethodGet {
		response := model.ResponseJson{Status: http.StatusMethodNotAllowed, Message: "Not allowed."}
		json.NewEncoder(w).Encode(response)
		return
	}
	if r.URL.Path != "/" {
		response := model.ResponseJson{Status: http.StatusNotFound, Message: "Not found."}
		json.NewEncoder(w).Encode(response)
		return
	}
	if r.Body == nil {
		response := model.ResponseJson{Status: http.StatusBadRequest, Message: "Please send a request body."}
		json.NewEncoder(w).Encode(response)
		return
	}

	response := model.ResponseJson{Status: http.StatusOK, Message: "Success."}

	switch r.Method {
	case http.MethodPost:
		response = s.postHandle(r)
	case http.MethodGet:
		response = s.getHandle(r)
	}

	json.NewEncoder(w).Encode(response)
}

func (s Server) postHandle(r *http.Request) model.ResponseJson {
	contentType := r.Header.Get(contentTypeHeader)
	var response model.ResponseJson

	if strings.Contains(contentType, jsonContentType) {
		response = s.loadFromJson(r)
	} else if strings.Contains(contentType, multipartFormDataContentType) {
		response = s.loadFromForm(r)
	} else {
		response = model.ResponseJson{Message: "Invalid Content-Type.", Status: http.StatusBadRequest}
	}

	return response
}

func (s Server) getHandle(r *http.Request) model.ResponseJson {
	response := model.ResponseJson{Status: http.StatusOK, Message: "Success."}
	url := r.URL.Query().Get("url")

	if len(url) == 0 {
		response = model.ResponseJson{Message: "Parameter missed.", Status: http.StatusBadRequest}
	}

	success := s.imageService.GetByUrl(url)

	if !success {
		response.Message = "Can't load by url: " + url
	}

	return response
}

func (s *Server) loadFromForm(r *http.Request) model.ResponseJson {
	response := model.ResponseJson{Status: http.StatusOK, Message: "Success."}
	err := r.ParseMultipartForm(32 << 20)

	if err != nil {
		glg.Errorf("Parse form", err.Error())
		response.Message = "Invalid form."
		return response
	}

	form := r.MultipartForm
	success, errors := s.imageService.GetFromMultipartFormData(form)

	if !success {
		response.Message = "Images with errors: " + strings.Join(errors, ", ")
	}

	return response
}

func (s Server) loadFromJson(r *http.Request) model.ResponseJson {
	response := model.ResponseJson{Status: http.StatusOK, Message: "Success."}
	postJson, err := getPostJsonData(r)

	if err != nil {
		glg.Errorf("Json parse", err.Error())
		return model.ResponseJson{Status: http.StatusBadRequest, Message: "Invalid JSON."}
	}

	for _, item := range postJson.Images {
		success := s.imageService.DecodeFromBase64(item.Base64, item.Format)

		if !success {
			response.Message = "Decode impossible"
		}
	}

	return response
}

func getPostJsonData(r *http.Request) (*model.RequestJson, error) {
	result := new(model.RequestJson)
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(result)

	if err != nil {
		return nil, err
	}

	return result, nil
}
