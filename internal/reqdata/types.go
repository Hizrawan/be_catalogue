package reqdata

import "io"

// UploadedFile is a struct representing a file uploaded by the user
type UploadedFile struct {
	Filename    string
	Size        int64
	ContentType string
	File        io.Reader
}

func (u *UploadedFile) Empty() bool {
	return u.File == nil
}
