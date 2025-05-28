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

func TestConsistencyPNG(t *testing.T) {
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

func TestConsistencyWEBP(t *testing.T) {
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

	webp1, _, err := img1.ExportWebp(&vips.WebpExportParams{})
	if err != nil {
		t.Fatalf("Failed to export first image: %v", err)
	}

	webp2, _, err := img2.ExportWebp(&vips.WebpExportParams{})
	if err != nil {
		t.Fatalf("Failed to export second image: %v", err)
	}

	hash1 := md5Hash(webp1)
	hash2 := md5Hash(webp2)

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
