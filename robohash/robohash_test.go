package robohash

import (
	"bytes"
	"crypto/md5"
	"encoding/hex"
	"image/png"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/Kagami/go-avif"
)

type testCase struct {
	name          string
	text          string
	set           string
	size          string
	bgSet         string
	png_expected  string
	avif_expected string
}

// getTestAssetsPath возвращает правильный путь к тестовым ассетам
func getTestAssetsPath() string {
	_, filename, _, _ := runtime.Caller(0)
	return filepath.Join(filepath.Dir(filename), "..", "assets")
}

func TestRoboHashGeneration(t *testing.T) {

	// Устанавливаем тестовые ассеты
	originalAssetsDir := assetsDir
	assetsDir = getTestAssetsPath()
	defer func() { assetsDir = originalAssetsDir }()

	tests := []testCase{
		{
			name:          "Default set1 with simple text",
			text:          "test123",
			set:           "set1",
			size:          "300x300",
			bgSet:         "",
			png_expected:  "9ac002dbb9998cf224e405f008fd5964",
			avif_expected: "eb0f46ed143b53f93fe565cee31c7d7a",
		},
		{
			name:          "set2 with different text",
			text:          "another_test",
			set:           "set2",
			size:          "350x350",
			bgSet:         "",
			png_expected:  "0ed912cb77b254cc2ef646bbc5329f4b",
			avif_expected: "99a5c32d91d826abae7bcd98a2411731",
		},
		{
			name:          "set3 with background",
			text:          "complex_robot",
			set:           "set3",
			size:          "500x500",
			bgSet:         "bg1",
			png_expected:  "fe4754ae3a9f8e5d42a8ff01b8379f74",
			avif_expected: "e4fb763d94b049e6c1a4e906c0fe6fa0",
		},
		{
			name:          "set4 with custom size",
			text:          "cat_avatar",
			set:           "set4",
			size:          "200x200",
			bgSet:         "",
			png_expected:  "42a8b00b31e43297dfdf36bec71bda69",
			avif_expected: "759857ef30c46a51d8f77ccd3a202d16",
		},
		{
			name:          "set5 human avatar",
			text:          "human_user",
			set:           "set5",
			size:          "400x400",
			bgSet:         "bg2",
			png_expected:  "4bbefc72b93c3b33de16e5179d2dca8c",
			avif_expected: "38d71b24f639c219941ce1634ecbfe43",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			robo := NewRoboHash(tt.text, tt.set)
			robo.Size = tt.size
			robo.BGSet = tt.bgSet

			img, err := robo.Generate()
			if err != nil {
				t.Fatalf("Generate() failed: %v", err)
			}
			if img == nil {
				t.Fatalf("Generate() failed, generated image is nil")
			}
			test_img_buf := new(bytes.Buffer)

			if err := png.Encode(test_img_buf, img); err != nil {
				t.Fatalf("encode to PNG failed %v", err)
			}

			hash := md5HashImageBuf(test_img_buf)
			if hash != tt.png_expected {
				t.Errorf("PNG image MD5 hash mismatch:\nGot: %s\nExpected: %s", hash, tt.png_expected)
			}

			test_img_buf.Reset()

			if err := avif.Encode(test_img_buf, img, nil); err != nil {
				t.Fatalf("encode to PNG failed %v", err)
			}

			hash = md5HashImageBuf(test_img_buf)
			if hash != tt.avif_expected {
				t.Errorf("AVIF image MD5 hash mismatch:\nGot: %s\nExpected: %s", hash, tt.avif_expected)
			}

		})
	}
}

func TestEmptyText(t *testing.T) {

	// Устанавливаем тестовые ассеты
	originalAssetsDir := assetsDir
	assetsDir = getTestAssetsPath()
	defer func() { assetsDir = originalAssetsDir }()

	robo := NewRoboHash("", "set1")
	img, err := robo.Generate()
	if err != nil {
		t.Fatalf("Generate() with empty text failed: %v", err)
	}

	// Проверяем что получили какое-то изображение
	if img == nil {
		t.Error("Generated image is nil for empty text")
	}
}

func TestInvalidSet(t *testing.T) {

	// Устанавливаем тестовые ассеты
	originalAssetsDir := assetsDir
	assetsDir = getTestAssetsPath()
	defer func() { assetsDir = originalAssetsDir }()

	robo := NewRoboHash("test", "invalid_set")
	_, err := robo.Generate()
	if err == nil {
		t.Error("Expected error for invalid set, got nil")
	}
}

func md5HashImageBuf(buf *bytes.Buffer) string {
	hasher := md5.New()
	hasher.Write(buf.Bytes())
	return hex.EncodeToString(hasher.Sum(nil))
}
