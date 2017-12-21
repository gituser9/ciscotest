package model

type Config struct {
	Port                  int    `json:"port"`
	LogFilePath           string `json:"log_file_path"`
	ImageDirectory        string `json:"image_directory"`
	ResizedImageDirectory string `json:"resized_image_directory"`
	ImageWidth            int    `json:"image_width"`
	ImageHeight           int    `json:"image_height"`
}
