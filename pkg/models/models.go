package models

type File struct {
	Name         string   `json:"name"`
	FileCode     string   `json:"fileCode"`
	Exist        bool     `json:"exist"`
	URL          string   `json:"url"`
	QRcodeBase64 string   `json:"qrCodeBase64"`
	FileNameList []string `json:"fileNameList"`
	OneDownload  bool     `json:"oneDownload"`
}

type FileSize struct {
	Size int `json:"size"`
}
