package models

type File struct {
	Name            string   `json:"name"`
	FileCode        string   `json:"fileCode"`
	Exist           bool     `json:"exist"`
	URL             string   `json:"url"`
	QRcodeBase64    string   `json:"qrCodeBase64"`
	FileNameList    []string `json:"fileNameList"`
	OneTimeDownload bool     `json:"oneTimeDownload"`
	LifeTime        string   `json:"lifeTime"`
}

type FileSize struct {
	Size int `json:"size"`
}
