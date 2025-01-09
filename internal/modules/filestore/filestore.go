package filestore

import (
	"errors"
	"time"
)

type Visibility string

const (
	Private Visibility = "private"
	Public  Visibility = "public"
)

var ErrFileNotExist = errors.New("file not exist")
var ErrInvalidVisibility = errors.New("invalid visibility value")

type Disk interface {
	Exists(filepath string) (bool, error)

	GetFile(filepath string) (*StoredFile, error)
	WriteFile(filepath string, file []byte) (*StoredFile, error)
	ReadFile(filepath string) ([]byte, error)
	MoveFile(from string, to string) (string, error)
	DeleteFile(filepath string) error

	GetVisibility(filepath string) (Visibility, error)
	SetVisibility(filepath string, visibility Visibility) error
	MakePublic(filepath string) (err error)
	MakePrivate(filepath string) (err error)

	GetURL(filepath string) (string, error)
	GetSignedURL(filepath string, expires time.Time) (string, error)
}

type StoredFile struct {
	content     []byte
	disk        Disk
	path        string
	Filename    string
	Extension   string
	Size        int64
	Visibility  Visibility
	ContentType string
}

func (f *StoredFile) Read() ([]byte, error) {
	if f.content == nil {
		c, err := f.disk.ReadFile(f.path)
		if err != nil {
			return nil, err
		}
		f.content = c
	}
	return f.content, nil
}

func (f *StoredFile) Move(to string) error {
	to, err := f.disk.MoveFile(f.path, to)
	if err != nil {
		return err
	}
	f.path = to
	return nil
}

func (f *StoredFile) Delete() error {
	return f.disk.DeleteFile(f.path)
}

func (f *StoredFile) URL() (string, error) {
	return f.disk.GetURL(f.path)
}

func (f *StoredFile) SignedURL(expires time.Time) (string, error) {
	return f.disk.GetSignedURL(f.path, expires)
}

func (f *StoredFile) GetVisibility() (Visibility, error) {
	return f.disk.GetVisibility(f.path)
}

func (f *StoredFile) SetVisibility(status Visibility) error {
	return f.disk.SetVisibility(f.path, status)
}

func (f *StoredFile) MakePrivate() error {
	return f.SetVisibility(Private)
}

func (f *StoredFile) MakePublic() error {
	return f.SetVisibility(Public)
}
