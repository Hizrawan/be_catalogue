package filestore

import (
	"context"
	"errors"
	"io/ioutil"
	"net/url"
	"path"
	"strings"
	"time"

	"cloud.google.com/go/storage"
	"google.golang.org/api/option"
)

type GoogleCloudStorageDisk struct {
	ctx        context.Context
	client     *storage.Client
	Name       string
	ProjectID  string
	Bucket     string
	PathPrefix string
	Visibility Visibility
}

type GCSDiskConfig struct {
	Name              string
	ProjectID         string
	Bucket            string
	PathPrefix        string
	DefaultVisibility string
	KeyFilePath       string
	KeyFileJSON       string
}

func NewGCSDisk(config GCSDiskConfig) (Disk, error) {
	ctx := context.Background()
	var opt option.ClientOption
	if config.KeyFileJSON != "" {
		opt = option.WithCredentialsJSON([]byte(config.KeyFileJSON))
	} else if config.KeyFilePath != "" {
		opt = option.WithCredentialsFile(config.KeyFilePath)
	}
	client, err := storage.NewClient(ctx, opt)
	if err != nil {
		return nil, err
	}

	return &GoogleCloudStorageDisk{
		ctx:        ctx,
		client:     client,
		ProjectID:  config.ProjectID,
		Bucket:     config.Bucket,
		PathPrefix: config.PathPrefix,
		Visibility: Visibility(config.DefaultVisibility),
	}, nil
}

func (d GoogleCloudStorageDisk) getQualifiedPath(filepath string) string {
	return path.Join(d.PathPrefix, filepath)
}

func (d GoogleCloudStorageDisk) Exists(filepath string) (bool, error) {
	_, err := d.client.Bucket(d.Bucket).Object(d.getQualifiedPath(filepath)).Attrs(d.ctx)
	if err != nil && errors.Is(err, storage.ErrObjectNotExist) {
		return false, nil
	} else if err != nil {
		return false, err
	}
	return true, nil
}

func (d GoogleCloudStorageDisk) GetFile(filepath string) (*StoredFile, error) {
	attrs, err := d.client.Bucket(d.Bucket).Object(d.getQualifiedPath(filepath)).Attrs(d.ctx)
	if err != nil && errors.Is(err, storage.ErrObjectNotExist) {
		return nil, ErrFileNotExist
	} else if err != nil {
		return nil, err
	}
	return &StoredFile{
		disk:        d,
		path:        d.normalizePath(filepath),
		Filename:    path.Base(filepath),
		Extension:   path.Ext(filepath),
		Size:        attrs.Size,
		Visibility:  d.getVisibilityFromACL(attrs.ACL),
		ContentType: attrs.ContentType,
	}, nil
}

func (d GoogleCloudStorageDisk) WriteFile(filepath string, file []byte) (*StoredFile, error) {
	obj := d.client.Bucket(d.Bucket).Object(d.getQualifiedPath(filepath))
	wc := obj.NewWriter(d.ctx)
	if _, err := wc.Write(file); err != nil {
		return nil, err
	}
	if err := wc.Close(); err != nil {
		return nil, err
	}

	if d.Visibility == Public {
		if err := d.MakePublic(filepath); err != nil {
			return nil, err
		}
	}

	attrs, err := obj.Attrs(d.ctx)
	if err != nil {
		return nil, err
	}

	return &StoredFile{
		disk:        d,
		path:        d.normalizePath(filepath),
		Filename:    path.Base(filepath),
		Extension:   path.Ext(filepath),
		Size:        attrs.Size,
		Visibility:  d.getVisibilityFromACL(attrs.ACL),
		ContentType: attrs.ContentType,
	}, nil
}

func (d GoogleCloudStorageDisk) ReadFile(filepath string) ([]byte, error) {
	rc, err := d.client.Bucket(d.Bucket).Object(d.getQualifiedPath(filepath)).NewReader(d.ctx)
	if err != nil {
		return nil, err
	}
	file, err := ioutil.ReadAll(rc)
	if err != nil {
		return nil, err
	}
	if err := rc.Close(); err != nil {
		return nil, err
	}

	return file, nil
}

func (d GoogleCloudStorageDisk) MoveFile(from string, to string) (string, error) {
	src := d.client.Bucket(d.Bucket).Object(d.getQualifiedPath(from))
	dst := d.client.Bucket(d.Bucket).Object(d.getQualifiedPath(to))

	dst = dst.If(storage.Conditions{DoesNotExist: true})

	if _, err := dst.CopierFrom(src).Run(d.ctx); err != nil {
		return "", err
	}
	if err := src.Delete(d.ctx); err != nil {
		return "", err
	}
	return d.normalizePath(to), nil
}

func (d GoogleCloudStorageDisk) DeleteFile(filepath string) error {
	return d.client.Bucket(d.Bucket).Object(d.getQualifiedPath(filepath)).Delete(d.ctx)
}

func (d GoogleCloudStorageDisk) getVisibilityFromACL(rules []storage.ACLRule) Visibility {
	for _, rule := range rules {
		if rule.Entity == storage.AllUsers {
			return Public
		}
	}
	return Private
}

func (d GoogleCloudStorageDisk) GetVisibility(filepath string) (Visibility, error) {
	acl := d.client.Bucket(d.Bucket).Object(d.getQualifiedPath(filepath)).ACL()
	rules, err := acl.List(d.ctx)
	if err != nil {
		return "", err
	}
	return d.getVisibilityFromACL(rules), nil
}

func (d GoogleCloudStorageDisk) SetVisibility(filepath string, visibility Visibility) error {
	obj := d.client.Bucket(d.Bucket).Object(d.getQualifiedPath(filepath))
	switch visibility {
	case Public:
		return obj.ACL().Set(d.ctx, storage.AllUsers, storage.RoleReader)
	case Private:
		return obj.ACL().Delete(d.ctx, storage.AllUsers)
	}
	return ErrInvalidVisibility
}

func (d GoogleCloudStorageDisk) MakePublic(filepath string) (err error) {
	return d.SetVisibility(filepath, Public)
}

func (d GoogleCloudStorageDisk) MakePrivate(filepath string) (err error) {
	return d.SetVisibility(filepath, Private)
}

func (d GoogleCloudStorageDisk) GetURL(filepath string) (string, error) {
	u := url.URL{
		Scheme: "https",
		Host:   "storage.googleapis.com",
		Path:   path.Join(d.Bucket, d.PathPrefix, filepath),
	}

	return u.String(), nil
}

func (d GoogleCloudStorageDisk) GetSignedURL(filepath string, expires time.Time) (string, error) {
	u, err := d.client.Bucket(d.Bucket).SignedURL(d.getQualifiedPath(filepath), &storage.SignedURLOptions{
		Method:  "GET",
		Expires: expires,
	})
	if err != nil {
		return "", err
	}
	return u, nil
}

func (d GoogleCloudStorageDisk) normalizePath(filepath string) string {
	filepath = strings.TrimLeft(filepath, "/")
	return "/" + filepath
}
