package server

import (
	"bytes"
	"encoding/json"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"model"
)

var server *Server

func init() {
	currentDirectory, err := os.Getwd()

	if err != nil {
		panic("Get directory error")
	}

	cfg := new(model.Config)
	cfg.Port = 13000
	cfg.ImageHeight = 100
	cfg.ImageWidth = 100
	cfg.ImageDirectory = currentDirectory + "/images/"
	cfg.ResizedImageDirectory = currentDirectory + "/images_resize/"
	cfg.LogFilePath = currentDirectory + "/info.log"

	server = CreateServer(cfg)
}

func TestJsonRequestValid(t *testing.T) {
	body := strings.NewReader(`{
	 	"images": [{
	 		"format": "png",
	 		"base_64": "iVBORw0KGgoAAAANSUhEUgAAABAAAAAQAQMAAAAlPW0iAAAABlBMVEUAAAD///+l2Z/dAAAAM0lEQVR4nGP4/5/h/1+G/58ZDrAz3D/McH8yw83NDDeNGe4Ug9C9zwz3gVLMDA/A6P9/AFGGFyjOXZtQAAAAAElFTkSuQmCC"
	 	}]
	 }`)
	request := httptest.NewRequest(http.MethodPost, "http://127.0.0.1:13000", body)
	request.Header.Set(contentTypeHeader, jsonContentType)

	recorder := httptest.NewRecorder()
	server.handler(recorder, request)

	response := recorder.Result()

	if response.StatusCode != http.StatusOK {
		t.Error("invalid code")
	}
}

func TestJsonRequestInvalidJson(t *testing.T) {
	body := strings.NewReader(`{
	 	"images": [
	 		"format": "png",
	 		"base_64": "iVBORw0KGgoAAAANSUhEUgAAABAAAAAQAQMAAAAlPW0iAAAABlBMVEUAAAD///+l2Z/dAAAAM0lEQVR4nGP4/5/h/1+G/58ZDrAz3D/McH8yw83NDDeNGe4Ug9C9zwz3gVLMDA/A6P9/AFGGFyjOXZtQAAAAAElFTkSuQmCC"
	 	}]
	 }`)
	request := httptest.NewRequest(http.MethodPost, "/", body)
	request.Header.Set(contentTypeHeader, jsonContentType)

	recorder := httptest.NewRecorder()
	server.handler(recorder, request)

	var responseBody model.ResponseJson
	decoder := json.NewDecoder(recorder.Body)
	err := decoder.Decode(&responseBody)

	if err != nil {
		t.Fatal("Invalid response JSON")
	}

	if responseBody.Status == http.StatusOK && responseBody.Status != http.StatusBadRequest && responseBody.Message != "Invalid JSON." {
		t.Error("Invalid code")
	}
}

func TestJsonRequestInvalidContentType(t *testing.T) {
	body := strings.NewReader(`{
	 	"images": [{
	 		"format": "png",
	 		"base_64": "iVBORw0KGgoAAAANSUhEUgAAABAAAAAQAQMAAAAlPW0iAAAABlBMVEUAAAD///+l2Z/dAAAAM0lEQVR4nGP4/5/h/1+G/58ZDrAz3D/McH8yw83NDDeNGe4Ug9C9zwz3gVLMDA/A6P9/AFGGFyjOXZtQAAAAAElFTkSuQmCC"
	 	}]
	 }`)
	request := httptest.NewRequest(http.MethodPost, "/", body)
	request.Header.Set(contentTypeHeader, "text/html")

	recorder := httptest.NewRecorder()
	server.handler(recorder, request)

	var responseBody model.ResponseJson
	decoder := json.NewDecoder(recorder.Body)
	err := decoder.Decode(&responseBody)

	if err != nil {
		t.Fatal("Invalid response JSON")
	}

	if responseBody.Status == http.StatusOK && responseBody.Status != http.StatusBadRequest && responseBody.Message != "Invalid Content-Type." {
		t.Error("Invalid code")
	}
}

func TestRequestFormValid(t *testing.T) {
	path := "../../images_for_test/"
	files, err := ioutil.ReadDir(path)

	if err != nil {
		t.Fatal("Read dir with test images: ", err.Error())
	}

	for _, fileInfo := range files {
		file, err := os.Open(path + fileInfo.Name())

		if err != nil {
			t.Error("Open image error: ", err.Error())
		}

		var body bytes.Buffer
		writer := multipart.NewWriter(&body)
		fw, err := writer.CreateFormFile("images", fileInfo.Name())

		if _, err = io.Copy(fw, file); err != nil {
			t.Error("Copy file error:", err.Error())
		}

		file.Close()

		handler := func(w http.ResponseWriter, r *http.Request) {
			io.WriteString(w, "")
		}
		request := httptest.NewRequest(http.MethodPost, "/", &body)
		request.Header.Set("Content-Type", writer.FormDataContentType())
		writer.Close()

		recorder := httptest.NewRecorder()
		handler(recorder, request)
		response := recorder.Result()

		if response.StatusCode != http.StatusOK {
			t.Error("invalid code")
		}
	}
}

func TestRequestFormInvalid(t *testing.T) {
	request := httptest.NewRequest(http.MethodPost, "/", nil)
	request.Header.Set("Content-Type", multipartFormDataContentType)

	recorder := httptest.NewRecorder()
	server.handler(recorder, request)

	var responseBody model.ResponseJson
	decoder := json.NewDecoder(recorder.Body)
	err := decoder.Decode(&responseBody)

	if err != nil {
		t.Fatal("Invalid response JSON")
	}

	if responseBody.Status == http.StatusOK && responseBody.Message != "Invalid form." {
		t.Error("invalid code")
	}
}

func TestRequestUrlInvalid(t *testing.T) {
	request := httptest.NewRequest(http.MethodGet, "/", nil)
	recorder := httptest.NewRecorder()
	server.handler(recorder, request)

	var responseBody model.ResponseJson
	decoder := json.NewDecoder(recorder.Body)
	err := decoder.Decode(&responseBody)

	if err != nil {
		t.Fatal("Invalid response JSON")
	}

	if responseBody.Status == http.StatusOK && responseBody.Status != http.StatusBadRequest {
		t.Error("Invalid code")
	}
}

func TestBadMethods(t *testing.T) {
	badMethods := []string{
		http.MethodConnect,
		http.MethodDelete,
		http.MethodHead,
		http.MethodOptions,
		http.MethodPatch,
		http.MethodPut,
		http.MethodTrace,
	}
	for _, method := range badMethods {
		request := httptest.NewRequest(method, "/", nil)
		recorder := httptest.NewRecorder()
		server.handler(recorder, request)

		var responseBody model.ResponseJson
		decoder := json.NewDecoder(recorder.Body)
		err := decoder.Decode(&responseBody)

		if err != nil {
			t.Error("Invalid response JSON")
			continue
		}

		if responseBody.Status == http.StatusOK && responseBody.Status != http.StatusBadRequest {
			t.Error("Invalid code")
		}
	}
}

func TestBadPaths(t *testing.T) {
	badPaths := []string{"/path1", "/path/2", "/path/3/3"}

	for _, path := range badPaths {
		request := httptest.NewRequest(http.MethodGet, path, nil)
		recorder := httptest.NewRecorder()
		server.handler(recorder, request)

		var responseBody model.ResponseJson
		decoder := json.NewDecoder(recorder.Body)
		err := decoder.Decode(&responseBody)

		if err != nil {
			t.Error("Invalid response JSON")
			continue
		}

		if responseBody.Status == http.StatusOK && responseBody.Status != http.StatusBadRequest {
			t.Error("Invalid code")
		}
	}
}

func TestEmptyBody(t *testing.T) {
	request := httptest.NewRequest(http.MethodPost, "/", nil)
	recorder := httptest.NewRecorder()
	server.handler(recorder, request)

	var responseBody model.ResponseJson
	decoder := json.NewDecoder(recorder.Body)
	err := decoder.Decode(&responseBody)

	if err != nil {
		t.Fatal("Invalid response JSON")
	}

	if responseBody.Status == http.StatusOK && responseBody.Status != http.StatusBadRequest && responseBody.Message != "Please send a request body." {
		t.Error("Invalid code")
	}
}
