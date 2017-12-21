package model

// Request
type RequestJson struct {
	Images []EncodingImage `json:"images"`
}

type EncodingImage struct {
	Format string `json:"format"`
	Base64 string `json:"base_64"`
}

// Response
type ResponseJson struct {
	Status  int    `json:"status"`
	Message string `json:"message"`
}
