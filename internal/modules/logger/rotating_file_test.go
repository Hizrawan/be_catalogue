package logger

import (
	"os"
	"path"
	"strings"
	"testing"
	"time"
)

func TestRotatingFileWriter(t *testing.T) {
	tmpDir := t.TempDir()
	writer := NewRotatingFileWriter(tmpDir, "test", true)

	t.Run("WriteFile writes in correct location", func(t *testing.T) {
		now = func() time.Time {
			return time.Date(2009, 11, 17, 20, 34, 58, 651387237, time.UTC)
		}

		writer.Write("test", nil, nil)

		out := path.Join(tmpDir, "2009-11-17-test.log")

		_, err := os.Stat(out)
		if err != nil {
			t.Errorf("want %v; got %v", nil, err)
		}
	})

	t.Run("WriteFile rotate file when date change", func(t *testing.T) {
		now = func() time.Time {
			return time.Date(2009, 11, 17, 20, 34, 58, 651387237, time.UTC)
		}

		writer.Write("test", nil, nil)

		out := path.Join(tmpDir, "2009-11-17-test.log")
		if _, err := os.Stat(out); os.IsNotExist(err) {
			t.Errorf("file %v should exist", out)
		} else if err != nil {
			t.Errorf("want %v; got %v", nil, err)
		}

		newOut := path.Join(tmpDir, "2009-11-18-test.log")
		if _, err := os.Stat(newOut); os.IsExist(err) {
			t.Errorf("file %v should not exist yet", newOut)
		} else if err != nil && !os.IsNotExist(err) {
			t.Errorf("want %v; got %v", nil, err)
		}

		now = func() time.Time {
			return time.Date(2009, 11, 18, 20, 34, 58, 651387237, time.UTC)
		}

		writer.Write("test", nil, nil)

		if _, err := os.Stat(newOut); os.IsNotExist(err) {
			t.Errorf("file %v should exist", newOut)
		} else if err != nil {
			t.Errorf("want %v; got %v", nil, err)
		}
	})

	t.Run("Ensure correct log is written", func(t *testing.T) {
		now = func() time.Time {
			return time.Date(2009, 11, 17, 20, 34, 58, 651387237, time.UTC)
		}

		writer.Write("test random log", nil, nil)
		writer.Write("test two random log", nil, nil)

		out := path.Join(tmpDir, "2009-11-17-test.log")

		content, err := os.ReadFile(out)
		if err != nil {
			t.Errorf("want %v; got %v", nil, err)
		}

		if !strings.Contains(string(content), "test random log") {
			t.Errorf("want contain %v; got %v", "test random log", string(content))
		}

		if strings.Contains(string(content), "testing not exist") {
			t.Errorf("should not contain %v; got %v", "testing not exist", string(content))
		}
	})
}
