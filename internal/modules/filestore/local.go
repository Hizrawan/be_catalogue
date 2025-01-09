package filestore

import (
	"bytes"
	"encoding/gob"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"path"
	"strings"
	"time"

	"be20250107/internal/constants"

	"github.com/lestrrat-go/jwx/v2/jwa"
	"github.com/lestrrat-go/jwx/v2/jwk"
	"github.com/lestrrat-go/jwx/v2/jwt"
)

type LocalDisk struct {
	BaseURL        string
	Name           string
	Dir            string
	PathPrefix     string
	SigningKey     jwk.RSAPrivateKey
	EnableMetadata bool
}

type LocalDiskConfig struct {
	BaseURL    string
	Name       string
	Dir        string
	PathPrefix string
	SigningKey jwk.RSAPrivateKey
}

type FileMetadata struct {
	Visibility Visibility
}

func NewLocalDisk(cfg LocalDiskConfig, enableMetadata bool) (Disk, error) {
	p := path.Join(cfg.Dir, cfg.PathPrefix)
	if s, err := os.Stat(p); err != nil && errors.Is(err, os.ErrNotExist) {
		err := os.MkdirAll(p, 0775)
		if err != nil {
			panic(err.Error())
		}
	} else if !s.IsDir() {
		panic(fmt.Errorf("a file exists on %s instead of a directory", p))
	}
	return &LocalDisk{
		BaseURL:        cfg.BaseURL,
		Name:           cfg.Name,
		Dir:            cfg.Dir,
		PathPrefix:     cfg.PathPrefix,
		EnableMetadata: enableMetadata,
		SigningKey:     cfg.SigningKey,
	}, nil
}

func (d LocalDisk) getQualifiedPath(filepath string) string {
	return path.Join(d.Dir, d.PathPrefix, filepath)
}

func (d LocalDisk) Exists(filepath string) (bool, error) {
	if _, err := os.Stat(d.getQualifiedPath(filepath)); err != nil && errors.Is(err, os.ErrNotExist) {
		return false, nil
	} else if err != nil {
		return false, err
	}
	return true, nil
}

func (d LocalDisk) GetFile(filepath string) (*StoredFile, error) {
	stat, err := os.Stat(d.getQualifiedPath(filepath))
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, ErrFileNotExist
		}
		return nil, err
	}

	f, err := os.Open(d.getQualifiedPath(filepath))
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = f.Close()
	}()
	buffer := make([]byte, 512)
	_, err = f.Read(buffer)
	if err != nil {
		return nil, err
	}

	v := Public
	if d.EnableMetadata {
		v, err = d.GetVisibility(filepath)
		if err != nil {
			return nil, err
		}
	}

	return &StoredFile{
		disk:        d,
		path:        d.normalizePath(filepath),
		Filename:    path.Base(filepath),
		Extension:   path.Ext(filepath),
		Size:        stat.Size(),
		Visibility:  v,
		ContentType: http.DetectContentType(buffer),
	}, nil
}

func (d LocalDisk) WriteFile(filepath string, file []byte) (*StoredFile, error) {
	q := d.getQualifiedPath(filepath)
	qp := path.Dir(q)
	_, err := os.Stat(qp)
	if err != nil && errors.Is(err, os.ErrNotExist) {
		err := os.MkdirAll(qp, 0775)
		if err != nil {
			return nil, err
		}
	} else if err != nil {
		return nil, err
	}
	if err := os.WriteFile(q, file, 0664); err != nil {
		return nil, err
	}

	stat, err := os.Stat(d.getQualifiedPath(filepath))
	if err != nil {
		return nil, err
	}

	v := Public
	if d.EnableMetadata {
		v, err = d.GetVisibility(filepath)
		if err != nil {
			return nil, err
		}
	}

	return &StoredFile{
		disk:        d,
		path:        d.normalizePath(filepath),
		Filename:    path.Base(filepath),
		Extension:   path.Ext(filepath),
		Size:        stat.Size(),
		Visibility:  v,
		ContentType: http.DetectContentType(file),
	}, nil
}

func (d LocalDisk) ReadFile(filepath string) ([]byte, error) {
	file, err := os.ReadFile(d.getQualifiedPath(filepath))
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, ErrFileNotExist
		}
		return nil, err
	}
	return file, nil
}

func (d LocalDisk) MoveFile(from string, to string) (string, error) {
	if err := os.Rename(d.getQualifiedPath(from), d.getQualifiedPath(to)); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return "", ErrFileNotExist
		}
	}
	return d.normalizePath(to), nil
}

func (d LocalDisk) DeleteFile(filepath string) error {
	return os.Remove(d.getQualifiedPath(filepath))
}

func (d LocalDisk) GetVisibility(filepath string) (Visibility, error) {
	if d.EnableMetadata {
		if f, err := os.ReadFile(d.getQualifiedPath(filepath + ".meta")); err != nil {
			if errors.Is(err, os.ErrNotExist) {
				return Public, nil
			}
			return "", err
		} else {
			var meta FileMetadata
			buf := bytes.NewBuffer(f)
			err = gob.NewDecoder(buf).Decode(&meta)
			if err != nil {
				return "", err
			}
			return meta.Visibility, nil
		}
	}
	return Public, nil
}

func (d LocalDisk) SetVisibility(filepath string, visibility Visibility) error {
	if d.EnableMetadata {
		p := path.Join(d.Dir, d.PathPrefix, filepath+".meta")
		var meta FileMetadata
		if f, err := os.ReadFile(p); err != nil {
			if errors.Is(err, os.ErrNotExist) {
				meta = FileMetadata{}
			} else {
				return err
			}
		} else {
			var meta FileMetadata
			buf := bytes.NewBuffer(f)
			err = gob.NewDecoder(buf).Decode(&meta)
			if err != nil {
				return err
			}
		}
		meta.Visibility = visibility
		var buf bytes.Buffer
		if err := gob.NewEncoder(&buf).Encode(meta); err != nil {
			return err
		}
		if err := os.WriteFile(p, buf.Bytes(), 0664); err != nil {
			return err
		}
	}
	return nil
}

func (d LocalDisk) MakePublic(filepath string) (err error) {
	return d.SetVisibility(filepath, Public)
}

func (d LocalDisk) MakePrivate(filepath string) (err error) {
	return d.SetVisibility(filepath, Private)
}

func (d LocalDisk) GetURL(filepath string) (string, error) {
	u, err := url.Parse(d.BaseURL)
	if err != nil {
		return "", err
	}
	u.Path = path.Join(u.Path, "/storage", d.Name, d.PathPrefix, filepath)
	return u.String(), nil
}

func (d LocalDisk) GetSignedURL(filepath string, expires time.Time) (string, error) {
	exists, err := d.Exists(filepath)
	if err != nil {
		return "", err
	}

	if !exists {
		return "", ErrFileNotExist
	}

	if !d.EnableMetadata {
		return d.GetURL(filepath)
	}

	payload, err := jwt.
		NewBuilder().
		Issuer(fmt.Sprintf("%s-Disk:%s", constants.TokenIssuer, d.Name)).
		Expiration(expires).
		IssuedAt(time.Now()).
		Subject(path.Join(d.PathPrefix, filepath)).
		Claim("d", d.Name).
		Build()
	if err != nil {
		return "", err
	}

	sign, err := jwt.Sign(payload, jwt.WithKey(jwa.RS256, d.SigningKey))
	if err != nil {
		return "", err
	}

	u, err := url.Parse(d.BaseURL)
	if err != nil {
		return "", err
	}
	u.Path = path.Join(u.Path, "/storage", d.Name, d.PathPrefix, filepath)
	q := u.Query()
	q.Set(constants.SignedURLSignatureQuery, string(sign))
	u.RawQuery = q.Encode()
	return u.String(), nil
}

func (d LocalDisk) TrimPrefix(filepath string) string {
	normalizedPrefix := d.normalizePath(d.PathPrefix)
	normalizedPath := d.normalizePath(filepath)
	return strings.TrimPrefix(normalizedPath, normalizedPrefix)
}

func (d LocalDisk) normalizePath(filepath string) string {
	filepath = strings.TrimLeft(filepath, "/")
	return "/" + filepath
}
