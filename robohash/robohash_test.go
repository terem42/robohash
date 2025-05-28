package robohash

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"strconv"
	"testing"

	"github.com/davidbyttow/govips/v2/vips"
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

func setupTests() {
	assetsDir = "../assets"
	vips.Startup(&vips.Config{
		ConcurrencyLevel: 0,
		MaxCacheFiles:    100,
		MaxCacheMem:      50 * 1024 * 1024,
		MaxCacheSize:     100,
		ReportLeaks:      true,
		CacheTrace:       false,
		CollectStats:     false,
	})
	vips.LoggingSettings(nil, vips.LogLevelWarning)
}

func TestMain(m *testing.M) {
	setupTests()
	m.Run()
}

func TestRoboHashGeneration(t *testing.T) {
	tests := []testCase{
		{
			name:          "Default set1 with simple text",
			text:          "test123",
			set:           "set1",
			size:          "300x300",
			bgSet:         "",
			png_expected:  "0d387613c5e8906ead77f9c721f72605",
			avif_expected: "5654cc45d232fb18550f79785fd2a40b",
			webp_expected: "c2c07051e15d01fe90a86c287503915f",
		},
		{
			name:          "set2 with different text",
			text:          "another_test",
			set:           "set2",
			size:          "350x350",
			bgSet:         "",
			png_expected:  "90bec5f6836c11aebf9655af401910d2",
			avif_expected: "bc1b9a1066bdfe8ec2fa71b9fefa7f5e",
			webp_expected: "2d097c23f664b8df7452f4d685368860",
		},
		{
			name:          "set3 with background",
			text:          "complex_robot",
			set:           "set3",
			size:          "500x500",
			bgSet:         "bg1",
			png_expected:  "d6d0fe605d5782f6754a67c3e344d12f",
			avif_expected: "2a690c0de0b15fe4483c65f2d2197406",
			webp_expected: "ee1de041d3bcd1eada710136ebe8c028",
		},
		{
			name:          "set4 with custom size",
			text:          "cat_avatar",
			set:           "set4",
			size:          "200x200",
			bgSet:         "",
			png_expected:  "3199253e1c92e2787fdd4e80c8c84583",
			avif_expected: "c38ebb3a1b0009fb139f796e40938fe0",
			webp_expected: "c031f034f2d40b1d4dead11fd8dc07d6",
		},
		{
			name:          "set5 human avatar",
			text:          "human_user",
			set:           "set5",
			size:          "400x400",
			bgSet:         "bg2",
			png_expected:  "c621edf3f80cc2a347d9de7c7a05b37a",
			avif_expected: "80ca3ac6a9cfd1b5f24d6a22eba0ab9e",
			webp_expected: "c7a401203a4a9b3df369e2ccf04c97bd",
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
			defer img.Close()

			pngBuf, _, err := img.ExportPng(&vips.PngExportParams{
				Quality: 100,
			})
			if err != nil {
				t.Fatalf("PNG export failed: %v", err)
			}
			imgHash := md5Hash(pngBuf)
			if imgHash != tt.png_expected {
				t.Errorf("PNG hash mismatch: got %s, want %s", imgHash, tt.png_expected)
			}

			avifBuf, _, err := img.ExportAvif(&vips.AvifExportParams{
				Quality: 100,
			})
			if err != nil {
				t.Fatalf("AVIF export failed: %v", err)
			}
			imgHash = md5Hash(avifBuf)
			if imgHash != tt.avif_expected {
				t.Errorf("AVIF hash mismatch: got %s, want %s", imgHash, tt.avif_expected)
			}

			webpBuf, _, err := img.ExportWebp(&vips.WebpExportParams{
				Quality:  100,
				Lossless: true,
			})
			if err != nil {
				t.Fatalf("WEBP export failed: %v", err)
			}
			imgHash = md5Hash(webpBuf)
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
	} else {
		defer img.Close()
	}
}

func TestInvalidSet(t *testing.T) {
	robo := NewRoboHash("test", "invalid_set")
	img, err := robo.Generate()
	if err == nil {
		t.Error("Expected error for invalid set, got nil")
		if img != nil {
			img.Close()
		}
	}
}

func TestAnySet(t *testing.T) {
	robo := NewRoboHash("test_any", "any")
	img, err := robo.Generate()
	if err != nil {
		t.Fatalf("Generate() with 'any' set failed: %v", err)
	}
	if img == nil {
		t.Error("Generated image is nil for 'any' set")
	} else {
		defer img.Close()

		width := img.Width()
		height := img.Height()

		validSizes := [][2]int{
			{300, 300},   // set1
			{350, 350},   // set2
			{1015, 1015}, // set3
			{1024, 1024}, // set4, set5
		}

		found := false
		for _, size := range validSizes {
			if width == size[0] && height == size[1] {
				found = true
				break
			}
		}

		if !found {
			t.Errorf("Unexpected image dimensions for 'any' set: %dx%d", width, height)
		}
	}
}

func TestBackgroundAny(t *testing.T) {
	robo := NewRoboHash("test_bg", "set1")
	robo.BGSet = "any"

	img, err := robo.Generate()
	if err != nil {
		t.Fatalf("Generate() with 'any' background failed: %v", err)
	}
	if img == nil {
		t.Error("Generated image is nil for 'any' background")
	} else {
		defer img.Close()
	}
}

func TestDifferentSizes(t *testing.T) {
	robo := &RoboHash{
		Text:  "example",
		Set:   "set1",
		Size:  "",
		BGSet: "",
	}

	sizes := []string{"100x100", "200x200", "400x400", "800x800"}

	for _, size := range sizes {
		t.Run("Size_"+size, func(t *testing.T) {
			robo.Size = size

			img, err := robo.Generate()
			if err != nil {
				t.Fatalf("Generate() with size %s failed: %v", size, err)
			}
			if img == nil {
				t.Error("Generated image is nil")
				return
			}

			var expectedWidth, expectedHeight int
			if _, err := fmt.Sscanf(size, "%dx%d", &expectedWidth, &expectedHeight); err != nil {
				t.Fatalf("Failed to parse size %s: %v", size, err)
			}

			if img.Width() != expectedWidth || img.Height() != expectedHeight {
				img.Close()
				t.Errorf("Size mismatch: expected %dx%d, got %dx%d",
					expectedWidth, expectedHeight, img.Width(), img.Height())
			}
			img.Close()
		})
	}
}

func TestConsistency(t *testing.T) {
	text := "consistency_test"
	set := "set1"

	robo1 := NewRoboHash(text, set)
	img1, err := robo1.Generate()
	if err != nil {
		t.Fatalf("First generation failed: %v", err)
	}
	defer img1.Close()

	robo2 := NewRoboHash(text, set)
	img2, err := robo2.Generate()
	if err != nil {
		t.Fatalf("Second generation failed: %v", err)
	}
	defer img2.Close()

	png1, _, err := img1.ExportPng(&vips.PngExportParams{Quality: 100})
	if err != nil {
		t.Fatalf("Failed to export first image: %v", err)
	}

	png2, _, err := img2.ExportPng(&vips.PngExportParams{Quality: 100})
	if err != nil {
		t.Fatalf("Failed to export second image: %v", err)
	}

	hash1 := md5Hash(png1)
	hash2 := md5Hash(png2)

	if hash1 != hash2 {
		t.Errorf("Images are not consistent: hash1=%s, hash2=%s", hash1, hash2)
	}
}

func TestAllSets(t *testing.T) {
	sets := []string{"set1", "set2", "set3", "set4", "set5"}

	for _, set := range sets {
		t.Run("Set_"+set, func(t *testing.T) {
			expectedWidth, expectedHeight := getSetDimensions(set)
			robo := NewRoboHash("test_"+set, set)
			robo.Size = strconv.Itoa(expectedWidth) + "x" + strconv.Itoa(expectedHeight)
			img, err := robo.Generate()
			if err != nil {
				t.Fatalf("Generate() for %s failed: %v", set, err)
			}
			if img == nil {
				t.Errorf("Generated image is nil for set %s", set)
				return
			}
			defer img.Close()

			// Verify image has expected dimensions for each set
			if img.Width() != expectedWidth || img.Height() != expectedHeight {
				t.Errorf("Wrong dimensions for %s: expected %dx%d, got %dx%d",
					set, expectedWidth, expectedHeight, img.Width(), img.Height())
			}
		})
	}
}

// Benchmark tests
func BenchmarkGenerate(b *testing.B) {
	robo := NewRoboHash("benchmark_test", "set1")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		img, err := robo.Generate()
		if err != nil {
			b.Fatalf("Generate failed: %v", err)
		}
		if img != nil {
			img.Close()
		}
	}
}

func BenchmarkGenerateWithResize(b *testing.B) {
	robo := NewRoboHash("benchmark_resize", "set1")
	robo.Size = "512x512"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		img, err := robo.Generate()
		if err != nil {
			b.Fatalf("Generate failed: %v", err)
		}
		if img != nil {
			img.Close()
		}
	}
}

func BenchmarkGenerateWithBackground(b *testing.B) {
	robo := NewRoboHash("benchmark_bg", "set1")
	robo.BGSet = "bg1"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		img, err := robo.Generate()
		if err != nil {
			b.Fatalf("Generate failed: %v", err)
		}
		if img != nil {
			img.Close()
		}
	}
}

func md5Hash(data []byte) string {
	hasher := md5.New()
	hasher.Write(data)
	return hex.EncodeToString(hasher.Sum(nil))
}
