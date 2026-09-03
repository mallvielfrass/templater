package models

type FileInfo struct {
	Sheets   []Sheet `json:"sheets"`
	FileName string  `json:"file_name"`
	Size     int     `json:"size"`
}
