package main

import (
	"encoding/json"
	"flag"
	"io/ioutil"
	"os"
	"strings"

	"model"
	"server"

	"github.com/kpango/glg"
)

var config *model.Config

const defaultConfigPath = "./cfg.json"

func init() {
	pathPtr := flag.String("config", defaultConfigPath, "Path for configuration file")
	flag.Parse()

	if *pathPtr == "" {
		panic("No config path")
	}

	bytes, err := ioutil.ReadFile(*pathPtr)

	if err != nil {
		// set default values
		config = new(model.Config)
		currentDirectory, err := os.Getwd()

		if err != nil {
			panic("Get directory error")
		}

		config.Port = 8080
		config.ImageHeight = 100
		config.ImageWidth = 100
		config.ImageDirectory = currentDirectory + "/images/"
		config.ResizedImageDirectory = currentDirectory + "/images_resize/"
		config.LogFilePath = currentDirectory + "/info.log"
	} else {
		json.Unmarshal(bytes, &config)
	}
	if !strings.HasSuffix(config.ImageDirectory, "/") {
		config.ImageDirectory += "/"
	}
	if !strings.HasSuffix(config.ResizedImageDirectory, "/") {
		config.ResizedImageDirectory += "/"
	}
	if _, err := os.Stat(config.ResizedImageDirectory); os.IsNotExist(err) {
		os.Mkdir(config.ResizedImageDirectory, 0777)
	}
	if _, err := os.Stat(config.ImageDirectory); os.IsNotExist(err) {
		os.Mkdir(config.ImageDirectory, 0777)
	}
}

func main() {
	logger := glg.FileWriter(config.LogFilePath, 0666)
	defer logger.Close()

	glg.Get().
		SetMode(glg.BOTH).
		AddLevelWriter(glg.ERR, logger) // log in file for errors only

	httpServer := server.CreateServer(config)
	httpServer.Start()
}
