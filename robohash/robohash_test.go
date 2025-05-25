package robohash

import (
	"bytes"
	"crypto/md5"
	"encoding/hex"
	"image/png"
	"testing"

	"github.com/Kagami/go-avif"
	"github.com/chai2010/webp"
)

type testCase struct {
	name          string
	text          string
	set           string
	size          string
	bgSet         string
	png_expected  string
	avif_expected string
	webp_expected string
}

func TestRoboHashGeneration(t *testing.T) {
	tests := []testCase{
		{
			name:          "Default set1 with simple text",
			text:          "test123",
			set:           "set1",
			size:          "300x300",
			bgSet:         "",
			png_expected:  "3f6f32e7aeac1cae6c62600f6f879779",
			avif_expected: "e58f7e3392f04ae009945d6c864de07b",
			webp_expected: "f38f40dde6152be2d4a740fe51dc2ebd",
		},
		{
			name:          "set2 with different text",
			text:          "another_test",
			set:           "set2",
			size:          "350x350",
			bgSet:         "",
			png_expected:  "ae5d7352f49b55ff4af0a09048c0663e",
			avif_expected: "7e8f23372d80791c83d346306d7c8ee6",
			webp_expected: "8e89dd256c7b5ffc6769fbb46a5dabb2",
		},
		{
			name:          "set3 with background",
			text:          "complex_robot",
			set:           "set3",
			size:          "500x500",
			bgSet:         "bg1",
			png_expected:  "508d1f14512da60aa3ba9bd93f3937e3",
			avif_expected: "e1e14027059152c8af398881fc11d58b",
			webp_expected: "9bb4a79588921bc4eb3e663d5e84b057",
		},
		{
			name:          "set4 with custom size",
			text:          "cat_avatar",
			set:           "set4",
			size:          "200x200",
			bgSet:         "",
			png_expected:  "7cbc9d0fde39a9644d3322ab93c14106",
			avif_expected: "7975ef5c2b1162c469d22a855d9d051e",
			webp_expected: "51ab01e37530843b1f95ab0f82fc7eac",
		},
		{
			name:          "set5 human avatar",
			text:          "human_user",
			set:           "set5",
			size:          "400x400",
			bgSet:         "bg2",
			png_expected:  "997f188de3228e39616b1d154a1f257d",
			avif_expected: "0fbed7170d0f8cb6439e99b9224e7069",
			webp_expected: "5a04bef70337ef4977345a1ce035fd24",
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
				t.Fatal("Generate() returned nil image")
			}

			imgBuf := new(bytes.Buffer)
			if err := png.Encode(imgBuf, img); err != nil {
				t.Fatalf("PNG encode failed: %v", err)
			}
			imgHash := md5Hash(imgBuf.Bytes())
			if imgHash != tt.png_expected {
				t.Errorf("PNG hash mismatch: got %s, want %s", imgHash, tt.png_expected)
			}

			imgBuf.Reset()
			if err := avif.Encode(imgBuf, img, nil); err != nil {
				t.Fatalf("AVIF encode failed: %v", err)
			}
			imgHash = md5Hash(imgBuf.Bytes())
			if imgHash != tt.avif_expected {
				t.Errorf("AVIF hash mismatch: got %s, want %s", imgHash, tt.avif_expected)
			}

			imgBuf.Reset()
			if err := webp.Encode(imgBuf, img, &webp.Options{Lossless: true}); err != nil {
				t.Fatalf("WEBP encode failed: %v", err)
			}
			imgHash = md5Hash(imgBuf.Bytes())
			if imgHash != tt.webp_expected {
				t.Errorf("WEBP hash mismatch: got %s, want %s", imgHash, tt.webp_expected)
			}

		})
	}
}

func TestEmptyText(t *testing.T) {
	robo := NewRoboHash("", "set1")
	img, err := robo.Generate()
	if err != nil {
		t.Fatalf("Generate() with empty text failed: %v", err)
	}
	if img == nil {
		t.Error("Generated image is nil for empty text")
	}
}

func TestInvalidSet(t *testing.T) {
	robo := NewRoboHash("test", "invalid_set")
	_, err := robo.Generate()
	if err == nil {
		t.Error("Expected error for invalid set, got nil")
	}
}

func md5Hash(data []byte) string {
	hasher := md5.New()
	hasher.Write(data)
	return hex.EncodeToString(hasher.Sum(nil))
}
