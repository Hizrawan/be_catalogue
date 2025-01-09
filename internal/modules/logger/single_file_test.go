package logger

import (
	"os"
	"path"
	"strings"
	"testing"
)

func TestSingleFileWriter(t *testing.T) {
	tmpDir := t.TempDir()
	writer := NewSingleFileWriter(tmpDir, "test", true)

	t.Run("WriteFile writes in correct location", func(t *testing.T) {
		writer.Write("test", nil, nil)

		out := path.Join(tmpDir, "test.log")

		_, err := os.Stat(out)
		if err != nil {
			t.Errorf("want %v; got %v", nil, err)
		}
	})

	t.Run("Ensure correct log is written", func(t *testing.T) {
		writer.Write("test random log", nil, nil)
		writer.Write("test two random log", nil, nil)

		out := path.Join(tmpDir, "test.log")

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
