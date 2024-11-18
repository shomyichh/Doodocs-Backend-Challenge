package models

type FileDetails struct {
	FilePath string  `json:"file_path"`
	Size     float64 `json:"size"`
	MimeType string  `json:"mimetype"`
}

type ArchiveInfo struct {
	Filename    string        `json:"filename"`
	ArchiveSize float64       `json:"archive_size"`
	TotalSize   float64       `json:"total_size"`
	TotalFiles  int           `json:"total_files"`
	Files       []FileDetails `json:"files"`
}
