#Golang test

Simple server for load images.

##For local build
- add $GOPATH to $PATH
- go get github.com/constabulary/gb/...
- gb vendor restore
- gb build
- cd bin
- ./main -config /path/for/config

**Config Example**
```
{
  "port": 8888,
  "log_file_path": "/path/for/create/log/file",
  "image_directory": "/path/for/save/images/",
  "resized_image_directory": "/path/for/save/images/preview",
  "image_width": 100,
  "image_height": 100
}
```

##Data for request
- JSON for POST Request
```
{
	"images": [{
		"format": "png",
		"base_64": "iVBORw0KGgoAAAANSUhEUgAAABAAAAAQAQMAAAAlPW0iAAAABlBMVEUAAAD///+l2Z/dAAAAM0lEQVR4nGP4/5/h/1+G/58ZDrAz3D/McH8yw83NDDeNGe4Ug9C9zwz3gVLMDA/A6P9/AFGGFyjOXZtQAAAAAElFTkSuQmCC"
	}]
}
```
- url for GET Request
```
http://127.0.0.1:8888/?url=https%3A%2F%2Fimages7.alphacoders.com%2F671%2F671281.jpg
```