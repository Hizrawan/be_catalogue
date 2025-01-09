package filestore

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"encoding/gob"
	"errors"
	"fmt"
	"net/url"
	"os"
	"path"
	"testing"
	"time"

	"be20250107/internal/constants"

	"github.com/lestrrat-go/jwx/v2/jwa"
	"github.com/lestrrat-go/jwx/v2/jwk"
	"github.com/lestrrat-go/jwx/v2/jwt"
)

func TestLocalDisk(t *testing.T) {
	tmpDir := t.TempDir()
	pathPrefix := "/test"
	disk, err := NewLocalDisk(LocalDiskConfig{
		BaseURL:    "http://localhost",
		Name:       "local",
		Dir:        tmpDir,
		PathPrefix: pathPrefix,
		SigningKey: nil,
	}, true)
	if err != nil {
		t.Fatal(err)
	}

	t.Run("WriteFile writes in correct location", func(t *testing.T) {
		out := path.Join(tmpDir, pathPrefix, "subdir/file.txt")
		_, err := os.Stat(out)
		if err == nil || !errors.Is(err, os.ErrNotExist) {
			t.Errorf("want %v; got %v", os.ErrNotExist, err)
		}

		_, err = disk.WriteFile("subdir/file.txt", []byte(""))
		if err != nil {
			t.Fatal(err)
		}

		_, err = os.Stat(out)
		if err != nil {
			t.Errorf("want %v; got %v", nil, err)
		}
	})

	t.Run("WriteFile writes correct data", func(t *testing.T) {
		_, err = disk.WriteFile("write/ereshkigal.txt", []byte("Kur Ki Gal Irkalla"))
		if err != nil {
			t.Fatal(err)
		}

		f, err := os.ReadFile(path.Join(tmpDir, pathPrefix, "write/ereshkigal.txt"))
		if err != nil {
			t.Errorf("want %v; got %v", nil, err)
		}

		if string(f) != "Kur Ki Gal Irkalla" {
			t.Errorf("want %v; got %v", "Kur Ki Gal Irkalla", f)
		}
	})

	t.Run("DeleteFile removes correct file", func(t *testing.T) {
		dir := prepareTestDirectory(tmpDir, pathPrefix, "delete")
		out := path.Join(dir, "kama.txt")
		err := os.WriteFile(out, []byte("Kama Sammohana"), 0664)
		if err != nil {
			t.Fatal(err)
		}

		err = disk.DeleteFile("delete/kama.txt")
		if err != nil {
			t.Errorf("want %v; got %v", nil, err)
		}

		_, err = os.Stat(out)
		if !errors.Is(err, os.ErrNotExist) {
			t.Errorf("want %v; got %v", os.ErrNotExist, err)
		}
	})

	t.Run("SetVisibility creates metadata file if set to Private with metadata support enabled", func(t *testing.T) {
		out := path.Join(tmpDir, pathPrefix, "meta/file")
		_, err := os.Stat(out)
		if err == nil || !errors.Is(err, os.ErrNotExist) {
			t.Errorf("want %v; got %v", os.ErrNotExist, err)
		}

		_, err = disk.WriteFile("meta/file", []byte(""))
		if err != nil {
			t.Fatal(err)
		}
		err = disk.SetVisibility("meta/file", Private)
		if err != nil {
			t.Fatal(err)
		}

		_, err = os.Stat(out + ".meta")
		if err != nil {
			t.Errorf("want %v; got %v", nil, err)
		}
	})

	t.Run("GetVisibility reads metadata file if exist", func(t *testing.T) {
		p := prepareTestDirectory(tmpDir, pathPrefix, "meta")
		out := path.Join(p, "ishtar.txt")
		err := os.WriteFile(out, []byte("An Gal Ta Ki Gal Šè"), 0664)
		if err != nil {
			t.Fatal(err)
		}
		meta := FileMetadata{Visibility: Private}
		var buf bytes.Buffer
		if err := gob.NewEncoder(&buf).Encode(meta); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(out+".meta", buf.Bytes(), 0664); err != nil {
			t.Fatal(err)
		}

		vis, err := disk.GetVisibility("meta/ishtar.txt")
		if err != nil {
			t.Fatal(err)
		}

		if vis != Private {
			t.Errorf("want %q; got %q", Private, vis)
		}
	})

	t.Run("GetVisibility returns Public if no metadata file exist", func(t *testing.T) {
		p := prepareTestDirectory(tmpDir, pathPrefix, "meta")
		out := path.Join(p, "spishtar.txt")
		err := os.WriteFile(out, []byte("Edin Shugurra Quasar"), 0664)
		if err != nil {
			t.Fatal(err)
		}

		vis, err := disk.GetVisibility("meta/spishtar.txt")
		if err != nil {
			t.Fatal(err)
		}

		if vis != Public {
			t.Errorf("want %q; got %q", Public, vis)
		}
	})

	t.Run("GetURL returns correct URL", func(t *testing.T) {
		cases := []struct {
			BaseURL    string
			Name       string
			PathPrefix string
			Path       string
			Want       string
		}{
			{
				"http://localhost",
				"local",
				"",
				"saber/artoria",
				"http://localhost/storage/local/saber/artoria",
			},
			{
				"https://domain.test/path",
				"disk",
				"",
				"archer/ishtar.png",
				"https://domain.test/path/storage/disk/archer/ishtar.png",
			},
			{
				"http://domain.uruk:9000/",
				"lnc",
				"/divine/5-star",
				"lancer/ereshkigal.pdf",
				"http://domain.uruk:9000/storage/lnc/divine/5-star/lancer/ereshkigal.pdf",
			},
			{
				"ftp://127.0.0.1",
				"rider",
				"///",
				"europa",
				"ftp://127.0.0.1/storage/rider/europa",
			},
			{
				"https://chaldea.bb/sakura",
				"summon",
				"assassin",
				"3/kama.png",
				"https://chaldea.bb/sakura/storage/summon/assassin/3/kama.png",
			},
		}

		for _, c := range cases {
			t.Run(fmt.Sprintf("testing %s", c.Want), func(t *testing.T) {
				d, err := NewLocalDisk(LocalDiskConfig{
					BaseURL:    c.BaseURL,
					Name:       c.Name,
					Dir:        tmpDir,
					PathPrefix: c.PathPrefix,
					SigningKey: nil,
				}, true)
				if err != nil {
					t.Fatal(err)
				}

				u, err := d.GetURL(c.Path)
				if err != nil {
					t.Fatal(err)
				}

				if u != c.Want {
					t.Errorf("want %q; got %q", c.Want, u)
				}
			})
		}
	})

	t.Run("GetSignedURL returns GetURL if metadata support is disabled", func(t *testing.T) {
		k, err := rsa.GenerateKey(rand.Reader, 2048)
		if err != nil {
			t.Fatal(err)
		}
		sk, err := jwk.FromRaw(k)
		if err != nil {
			t.Fatal(err)
		}
		d, err := NewLocalDisk(LocalDiskConfig{
			BaseURL:    "https://chaldea.bb/",
			Name:       "local",
			Dir:        tmpDir,
			PathPrefix: "",
			SigningKey: sk.(jwk.RSAPrivateKey),
		}, false)
		if err != nil {
			t.Fatal(err)
		}
		_, err = d.WriteFile("meltryllis", []byte(""))
		if err != nil {
			t.Fatal(err)
		}

		u, err := d.GetURL("meltryllis")
		if err != nil {
			t.Fatal(err)
		}
		su, err := d.GetSignedURL("meltryllis", time.Now().Add(time.Hour))
		if err != nil {
			t.Fatal(err)
		}

		if u != su {
			t.Errorf("want %q; got %q", u, su)
		}
	})

	t.Run("GetSignedURL returns signed URL with correct properties", func(t *testing.T) {
		k, err := rsa.GenerateKey(rand.Reader, 2048)
		if err != nil {
			t.Fatal(err)
		}
		sk, err := jwk.FromRaw(k)
		if err != nil {
			t.Fatal(err)
		}
		pk, err := sk.(jwk.RSAPrivateKey).PublicKey()
		if err != nil {
			t.Fatal(err)
		}
		d, err := NewLocalDisk(LocalDiskConfig{
			BaseURL:    "https://chaldea.bb/",
			Name:       "local",
			Dir:        tmpDir,
			PathPrefix: "caster",
			SigningKey: sk.(jwk.RSAPrivateKey),
		}, true)
		if err != nil {
			t.Fatal(err)
		}

		_, err = d.WriteFile("/tamamo.png", []byte("mikoon~"))
		if err != nil {
			t.Fatal(err)
		}

		sUrl, err := d.GetSignedURL("/tamamo.png", time.Now().Add(time.Hour))
		if err != nil {
			t.Fatal(err)
		}

		u, err := url.Parse(sUrl)
		if err != nil {
			t.Fatal(err)
		}
		if u.Host != "chaldea.bb" {
			t.Errorf("want %q; got %q", "chaldea.bb", u.Host)
		}
		if u.Path != "/storage/local/caster/tamamo.png" {
			t.Errorf("want %q; got %q", "/storage/local/caster/tamamo.png", u.Path)
		}
		if !u.Query().Has(constants.SignedURLSignatureQuery) {
			t.Errorf("no %s query in the signed URL", constants.SignedURLSignatureQuery)
		}
		tok, err := jwt.ParseString(u.Query().Get(constants.SignedURLSignatureQuery), jwt.WithKey(jwa.RS256, pk.(jwk.RSAPublicKey)))
		if err != nil {
			t.Fatal(err)
		}
		if tok.Expiration().Before(time.Now()) {
			t.Errorf("token has expired")
		}
		if tok.Issuer() != fmt.Sprintf("%s-Disk:local", constants.TokenIssuer) {
			t.Errorf("want %q; got %q", fmt.Sprintf("%s-Disk:local", constants.TokenIssuer), tok.Issuer())
		}
		if tok.Subject() != "caster/tamamo.png" {
			t.Errorf("want %q; got %q", "caster/tamamo.png", tok.Subject())
		}
		if dval, ok := tok.Get("d"); !ok {
			t.Error("cannot retrieve d claim")
		} else if dval != "local" {
			t.Errorf("want %q; got %q", "local", dval)
		}
	})
}

func prepareTestDirectory(root string, prefix string, test string) string {
	p := path.Join(root, prefix, test)
	if s, err := os.Stat(p); err != nil && errors.Is(err, os.ErrNotExist) {
		err := os.MkdirAll(p, 0775)
		if err != nil {
			panic(err.Error())
		}
	} else if !s.IsDir() {
		panic(fmt.Errorf("a file exists on %s instead of a directory", p))
	}
	return p
}
