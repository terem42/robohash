package robohash

import (
	"crypto/sha512"
	"encoding/hex"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/davidbyttow/govips/v2/vips"
	"github.com/facette/natsort"
)

var assetsDir = "assets"

func init() {
	vips.Startup(&vips.Config{
		ConcurrencyLevel: 0,
		MaxCacheFiles:    300,
		MaxCacheMem:      50 * 1024 * 1024, // 50MB initial cache
		MaxCacheSize:     100,
		ReportLeaks:      false,
		CacheTrace:       false,
		CollectStats:     false,
	})
	vips.LoggingSettings(nil, vips.LogLevelWarning)
}

type RoboHash struct {
	Text  string
	Set   string
	Size  string
	BGSet string
}

func NewRoboHash(text string, set string) *RoboHash {
	return &RoboHash{
		Text:  text,
		Set:   set,
		Size:  "300x300",
		BGSet: "",
	}
}

func (r *RoboHash) Generate() (*vips.ImageRef, error) {
	if r.Set == "" {
		r.Set = "set1"
	}

	sha512 := sha512.New()
	sha512.Write([]byte(r.Text))
	hashBytes := sha512.Sum(nil)
	hashString := hex.EncodeToString(hashBytes)

	hashParts := splitHashIntoParts(hashString, 11)

	if r.Set == "any" {
		sets, err := os.ReadDir(filepath.Join(assetsDir))
		if err != nil {
			return nil, fmt.Errorf("failed to read sets directory: %v", err)
		}

		var availableSets []string
		for _, entry := range sets {
			if entry.IsDir() && strings.HasPrefix(entry.Name(), "set") {
				availableSets = append(availableSets, entry.Name())
			}
		}

		if len(availableSets) == 0 {
			return nil, fmt.Errorf("no valid sets found")
		}

		setIndex := hexToInt(hashParts[1]) % len(availableSets)
		r.Set = availableSets[setIndex]
	}

	parts := make(map[string]string)

	switch r.Set {
	case "set1":
		entries, err := os.ReadDir(filepath.Join(assetsDir, r.Set))
		if err != nil {
			return nil, fmt.Errorf("failed to read set1 directory: %v", err)
		}

		var colorDirs []string
		for _, entry := range entries {
			if entry.IsDir() {
				colorDirs = append(colorDirs, filepath.Join(r.Set, entry.Name()))
			}
		}

		colorIndex := hexToInt(hashParts[0]) % len(colorDirs)
		colorPath := colorDirs[colorIndex]
		color := filepath.Base(colorPath)

		parts["mouth"] = selectPart(hashParts[4], filepath.Join(r.Set, color, "000#Mouth"))
		parts["eyes"] = selectPart(hashParts[5], filepath.Join(r.Set, color, "001#Eyes"))
		parts["accessory"] = selectPart(hashParts[6], filepath.Join(r.Set, color, "002#Accessory"))
		parts["body"] = selectPart(hashParts[7], filepath.Join(r.Set, color, "003#01Body"))
		parts["face"] = selectPart(hashParts[8], filepath.Join(r.Set, color, "004#02Face"))

	case "set2":
		parts["body"] = selectPart(hashParts[4], filepath.Join(r.Set, "000#04Body"))
		parts["mouth"] = selectPart(hashParts[5], filepath.Join(r.Set, "001#Mouth"))
		parts["eyes"] = selectPart(hashParts[6], filepath.Join(r.Set, "002#Eyes"))
		parts["bodycolors"] = selectPart(hashParts[7], filepath.Join(r.Set, "003#02BodyColors"))
		parts["facecolors"] = selectPart(hashParts[8], filepath.Join(r.Set, "004#01FaceColors"))
		parts["nose"] = selectPart(hashParts[9], filepath.Join(r.Set, "005#Nose"))
		parts["face"] = selectPart(hashParts[10], filepath.Join(r.Set, "006#03Faces"))

	case "set3":
		parts["mouth"] = selectPart(hashParts[4], filepath.Join(r.Set, "000#07Mouth"))
		parts["wave"] = selectPart(hashParts[5], filepath.Join(r.Set, "001#02Wave"))
		parts["eyebrows"] = selectPart(hashParts[6], filepath.Join(r.Set, "002#05Eyebrows"))
		parts["eyes"] = selectPart(hashParts[7], filepath.Join(r.Set, "003#04Eyes"))
		parts["nose"] = selectPart(hashParts[8], filepath.Join(r.Set, "004#06Nose"))
		parts["base"] = selectPart(hashParts[9], filepath.Join(r.Set, "005#01BaseFace"))
		parts["antenna"] = selectPart(hashParts[10], filepath.Join(r.Set, "006#03Antenna"))

	case "set4":
		parts["body"] = selectPart(hashParts[4], filepath.Join(r.Set, "000#00body"))
		parts["fur"] = selectPart(hashParts[5], filepath.Join(r.Set, "001#01fur"))
		parts["eyes"] = selectPart(hashParts[6], filepath.Join(r.Set, "002#02eyes"))
		parts["mouth"] = selectPart(hashParts[7], filepath.Join(r.Set, "003#03mouth"))
		parts["accessory"] = selectPart(hashParts[8], filepath.Join(r.Set, "004#04accessories"))

	case "set5":
		parts["body"] = selectPart(hashParts[4], filepath.Join(r.Set, "000#Body"))
		parts["eyes"] = selectPart(hashParts[5], filepath.Join(r.Set, "001#Eye"))
		parts["eyebrow"] = selectPart(hashParts[6], filepath.Join(r.Set, "002#Eyebrow"))
		parts["mouth"] = selectPart(hashParts[7], filepath.Join(r.Set, "003#Mouth"))
		parts["cloth"] = selectPart(hashParts[8], filepath.Join(r.Set, "004#Cloth"))
		parts["facialhair"] = selectPart(hashParts[9], filepath.Join(r.Set, "005#FacialHair"))
		parts["top"] = selectPart(hashParts[10], filepath.Join(r.Set, "006#Top"))
		parts["accessories"] = selectPart(hashParts[11], filepath.Join(r.Set, "007#Accessories"))

	default:
		return nil, fmt.Errorf("unknown set: %s", r.Set)
	}

	bgSetHash := hashParts[3]
	if r.BGSet == "any" {
		bgSets, err := os.ReadDir(filepath.Join(assetsDir, "backgrounds"))
		if err != nil {
			return nil, fmt.Errorf("failed to read backgrounds directory: %v", err)
		}
		bgSetIndex := hexToInt(bgSetHash) % len(bgSets)
		r.BGSet = bgSets[bgSetIndex].Name()
	}

	return composeImage(parts, r.Size, r.BGSet, r.Set, hashString[0:12])
}

func selectPart(hashPart string, partPath string) string {
	dirPath := filepath.Join(assetsDir, partPath)

	entries, err := os.ReadDir(dirPath)
	if err != nil {
		log.Printf("Error reading directory %s: %v", dirPath, err)
		return ""
	}

	var matches []string
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".png") {
			fullPath := filepath.Join(dirPath, entry.Name())
			matches = append(matches, fullPath)
		}
	}

	natsort.Sort(matches)

	if len(matches) == 0 {
		log.Printf("No PNG files found in directory: %s", dirPath)
		return ""
	}

	index := hexToInt(hashPart) % len(matches)
	return matches[index]
}

func getSetDimensions(set string) (int, int) {
	switch set {
	case "set1":
		return 300, 300
	case "set2":
		return 350, 350
	case "set3":
		return 1015, 1015
	case "set4", "set5":
		return 1024, 1024
	default:
		return 300, 300
	}
}

func normalizeImage(img *vips.ImageRef) error {
	if img.Interpretation() != vips.InterpretationSRGB {
		if err := img.ToColorSpace(vips.InterpretationSRGB); err != nil {
			return fmt.Errorf("failed to convert to sRGB: %v", err)
		}
	}

	bands := img.Bands()
	if bands == 3 {
		if err := img.AddAlpha(); err != nil {
			return fmt.Errorf("failed to add alpha channel: %v", err)
		}
	} else if bands != 4 {
		return fmt.Errorf("unexpected number of bands: %d", bands)
	}

	return nil
}

func composeImage(parts map[string]string, size string, bgSet string, set string, bgSetHashPart string) (*vips.ImageRef, error) {
	width, height := getSetDimensions(set)

	base, err := vips.Black(width, height)
	if err != nil {
		return nil, fmt.Errorf("failed to create base image: %v", err)
	}

	if err := base.ToColorSpace(vips.InterpretationSRGB); err != nil {
		base.Close()
		return nil, fmt.Errorf("failed to set color space: %v", err)
	}

	if err := base.AddAlpha(); err != nil {
		base.Close()
		return nil, fmt.Errorf("failed to add alpha channel: %v", err)
	}

	if err := base.Linear([]float64{1, 1, 1, 0}, []float64{0, 0, 0, 0}); err != nil {
		base.Close()
		return nil, fmt.Errorf("failed to make image transparent: %v", err)
	}

	if bgSet != "" {
		bgDirPath := filepath.Join(assetsDir, "backgrounds", bgSet)

		entries, err := os.ReadDir(bgDirPath)
		if err != nil {
			log.Printf("Error reading background directory %s: %v", bgDirPath, err)
		} else {
			var bgFiles []string
			for _, entry := range entries {
				if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".png") {
					bgFiles = append(bgFiles, filepath.Join(bgDirPath, entry.Name()))
				}
			}

			if len(bgFiles) > 0 {
				bgIndex := hexToInt(bgSetHashPart) % len(bgFiles)

				bgImg, err := loadAndResizeImage(bgFiles[bgIndex], width, height)
				if err != nil {
					base.Close()
					return nil, fmt.Errorf("error loading background: %v", err)
				}

				if err := normalizeImage(bgImg); err != nil {
					base.Close()
					bgImg.Close()
					return nil, fmt.Errorf("error normalizing background: %v", err)
				}

				if err := base.Composite(bgImg, vips.BlendModeOver, 0, 0); err != nil {
					base.Close()
					bgImg.Close()
					return nil, fmt.Errorf("error compositing background: %v", err)
				}
				bgImg.Close()
			}
		}
	}

	order := getPartsOrder(set)

	for _, partType := range order {
		if partPath, ok := parts[partType]; ok && partPath != "" {
			partImg, err := loadAndResizeImage(partPath, width, height)
			if err != nil {
				log.Printf("Error loading part %s (%s): %v", partType, partPath, err)
				continue
			}

			if err := normalizeImage(partImg); err != nil {
				log.Printf("Error normalizing part %s: %v", partType, err)
				partImg.Close()
				continue
			}

			if err := base.Composite(partImg, vips.BlendModeOver, 0, 0); err != nil {
				log.Printf("Error compositing part %s: %v", partType, err)
			}
			partImg.Close()
		}
	}

	if size != "" {
		sizeParts := strings.Split(size, "x")
		if len(sizeParts) == 2 {
			targetWidth, err1 := strconv.Atoi(sizeParts[0])
			targetHeight, err2 := strconv.Atoi(sizeParts[1])

			if err1 == nil && err2 == nil && (targetWidth != width || targetHeight != height) {
				log.Printf("resize to=%v\n", size)
				resized, err := resizeImageOptimized(base, targetWidth, targetHeight)
				if err != nil {
					base.Close()
					return nil, err
				}
				log.Printf("resized width=%v, height=%v", resized.Width(), resized.Height())
				return resized, nil
			}
		}
	}

	return base, nil
}

func splitHashIntoParts(hash string, count int) []string {
	partLength := len(hash) / count
	parts := make([]string, count)
	for i := 0; i < count; i++ {
		start := i * partLength
		end := (i + 1) * partLength
		parts[i] = hash[start:end]
	}
	parts = append(parts, parts...)
	return parts
}

func hexToInt(hexStr string) int {
	num, err := strconv.ParseInt(hexStr, 16, 64)
	if err != nil {
		return 0
	}
	return int(num)
}

func getPartsOrder(set string) []string {
	switch set {
	case "set1":
		return []string{
			"body",      // 003#01Body
			"face",      // 004#02Face
			"eyes",      // 001#Eyes
			"mouth",     // 000#Mouth
			"accessory", // 002#Accessory
		}

	case "set2":
		return []string{
			"facecolors", // 004#01FaceColors
			"bodycolors", // 003#02BodyColors
			"face",       // 006#03Faces
			"body",       // 000#04Body
			"mouth",      // 001#Mouth
			"eyes",       // 002#Eyes
			"nose",       // 005#Nose
		}

	case "set3":
		return []string{
			"base",     // 005#01BaseFace
			"wave",     // 001#02Wave
			"antenna",  // 006#03Antenna
			"eyes",     // 003#04Eyes
			"eyebrows", // 002#05Eyebrows
			"nose",     // 004#06Nose
			"mouth",    // 000#07Mouth
		}

	case "set4":
		return []string{
			"body",      // 000#00body
			"fur",       // 001#01fur
			"eyes",      // 002#02eyes
			"mouth",     // 003#03mouth
			"accessory", // 004#04accessories
		}

	case "set5":
		return []string{
			"body",        // 000#Body
			"eyes",        // 001#Eye
			"eyebrow",     // 002#Eyebrow
			"mouth",       // 003#Mouth
			"cloth",       // 004#Cloth
			"facialhair",  // 005#FacialHair
			"top",         // 006#Top
			"accessories", // 007#Accessories
		}

	default:
		return []string{
			"body",
			"face",
			"eyes",
			"mouth",
			"accessory",
		}
	}
}

func loadAndResizeImage(path string, width, height int) (*vips.ImageRef, error) {
	img, err := vips.LoadImageFromFile(path, &vips.ImportParams{})
	if err != nil {
		return nil, err
	}

	if img.Width() != width || img.Height() != height {
		scale := float64(width) / float64(img.Width())
		if err := img.Resize(scale, vips.KernelLanczos3); err != nil {
			img.Close()
			return nil, err
		}
	}

	return img, nil
}

func resizeImageOptimized(img *vips.ImageRef, targetWidth, targetHeight int) (*vips.ImageRef, error) {
	currentWidth := img.Width()
	currentHeight := img.Height()

	if currentWidth == targetWidth && currentHeight == targetHeight {
		return img, nil
	}

	scale := float64(targetWidth) / float64(currentWidth)

	if err := img.Resize(scale, vips.KernelLanczos3); err != nil {
		return nil, fmt.Errorf("failed to resize image: %v", err)
	}

	newHeight := img.Height()
	if newHeight != targetHeight {
		if err := img.ExtractArea(0, 0, targetWidth, targetHeight); err != nil {
			return nil, fmt.Errorf("failed to extract area: %v", err)
		}
	}

	return img, nil
}
