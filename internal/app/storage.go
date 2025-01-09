package app

import (
	"fmt"

	"be20250107/internal/config"
	"be20250107/internal/modules/filestore"

	"github.com/lestrrat-go/jwx/v2/jwk"
)

func NewDisks(cfg config.StorageConfig, signingKey jwk.RSAPrivateKey, pubCfg *config.PublicConfig) (map[string]filestore.Disk, error) {
	disks := make(map[string]filestore.Disk)
	for _, c := range cfg.Disks {
		switch c.Driver {
		case "gcs":
			disk, err := filestore.NewGCSDisk(filestore.GCSDiskConfig{
				Name:              c.Name,
				ProjectID:         c.ProjectID,
				Bucket:            c.Bucket,
				PathPrefix:        c.GCSDriverConfig.PathPrefix,
				DefaultVisibility: c.Visibility,
				KeyFilePath:       c.KeyFilePath,
				KeyFileJSON:       c.KeyFileJSON,
			})
			if err != nil {
				panic(err.Error())
			}
			disks[c.Name] = disk
		case "local":
			dir := c.Dir
			if dir == "" {
				dir = fmt.Sprintf("./storage/%s", c.Name)
			}
			disk, err := filestore.NewLocalDisk(filestore.LocalDiskConfig{
				BaseURL:    pubCfg.AppURL,
				Name:       c.Name,
				Dir:        dir,
				SigningKey: signingKey,
			}, true)
			if err != nil {
				panic(err.Error())
			}
			disks[c.Name] = disk
		case "public":
			disk, err := filestore.NewLocalDisk(filestore.LocalDiskConfig{
				BaseURL:    pubCfg.AppURL,
				Name:       "public",
				Dir:        "./storage/public",
				SigningKey: signingKey,
			}, false)
			if err != nil {
				return nil, err
			}
			disks[c.Name] = disk
		default:
			return nil, fmt.Errorf("invalid storage driver: %s", c.Driver)
		}
	}
	return disks, nil
}
