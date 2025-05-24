package robohash

import (
	"crypto/sha512"
	"embed"
	"encoding/hex"
	"fmt"
	"image"
	"image/draw"
	"log"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/facette/natsort"

	draw2 "golang.org/x/image/draw"
)

//go:embed assets
var assetsFS embed.FS

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

func (r *RoboHash) Generate() (image.Image, error) {

	if r.Set == "" {
		r.Set = "set1"
	}

	sha512 := sha512.New()
	sha512.Write([]byte(r.Text))
	hashBytes := sha512.Sum(nil)
	hashString := hex.EncodeToString(hashBytes)

	hashParts := splitHashIntoParts(hashString, 11)

	if r.Set == "any" {
		sets, err := assetsFS.ReadDir(filepath.Join("assets"))
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
		entries, err := assetsFS.ReadDir(filepath.Join("assets", r.Set))
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
		bgSets, err := assetsFS.ReadDir(filepath.Join("assets", "backgrounds"))
		if err != nil {
			return nil, fmt.Errorf("failed to read backgrounds directory: %v", err)
		}
		bgSetIndex := hexToInt(bgSetHash) % len(bgSets)
		r.BGSet = bgSets[bgSetIndex].Name()
	}

	return composeImage(parts, r.Size, r.BGSet, r.Set, hashString[0:12])
}

func selectPart(hashPart string, partPath string) string {
	dirPath := filepath.Join("assets", partPath)

	entries, err := assetsFS.ReadDir(dirPath)
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

func composeImage(parts map[string]string, size string, bgSet string, set string, bgSetHashPart string) (image.Image, error) {
	width, height := getSetDimensions(set)
	base := image.NewRGBA(image.Rect(0, 0, width, height))

	if bgSet != "" {
		bgDirPath := filepath.Join("assets", "backgrounds", bgSet)

		entries, err := assetsFS.ReadDir(bgDirPath)
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
					return nil, fmt.Errorf("error loading background: %v", err)
				}
				draw.Draw(base, base.Bounds(), bgImg, image.Point{}, draw.Over)
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
			draw.Draw(base, base.Bounds(), partImg, image.Point{}, draw.Over)
		}
	}

	if size != "" {
		return resizeImage(base, size)
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

func loadAndResizeImage(path string, width, height int) (image.Image, error) {
	img, err := loadImage(path)
	if err != nil {
		return nil, err
	}

	if img.Bounds().Dx() == width && img.Bounds().Dy() == height {
		return img, nil
	}

	return resizeImage(img, fmt.Sprintf("%dx%d", width, height))
}

func resizeImage(img image.Image, size string) (image.Image, error) {
	sizeParts := strings.Split(size, "x")
	if len(sizeParts) != 2 {
		return nil, fmt.Errorf("invalid size format")
	}

	width, err := strconv.Atoi(sizeParts[0])
	if err != nil {
		return nil, err
	}

	height, err := strconv.Atoi(sizeParts[1])
	if err != nil {
		return nil, err
	}

	resized := image.NewRGBA(image.Rect(0, 0, width, height))
	draw2.ApproxBiLinear.Scale(resized, resized.Bounds(), img, img.Bounds(), draw.Over, nil)

	return resized, nil
}

func loadImage(path string) (image.Image, error) {
	file, err := assetsFS.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	img, _, err := image.Decode(file)
	if err != nil {
		return nil, err
	}

	if _, ok := img.(*image.RGBA); !ok {
		rgba := image.NewRGBA(img.Bounds())
		draw.Draw(rgba, rgba.Bounds(), img, image.Point{}, draw.Src)
		return rgba, nil
	}

	return img, nil
}
