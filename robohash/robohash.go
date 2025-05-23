package robohash

import (
	"crypto/sha1"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"image"
	"image/draw"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	draw2 "golang.org/x/image/draw"
)

var assetsDir = "assets"

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
		Size:  "300x300", // размер по умолчанию
		BGSet: "",        // фон по умолчанию прозрачный
	}
}

func (r *RoboHash) Generate() (image.Image, error) {

	if r.Set == "" {
		r.Set = "set1"
	}

	// 1. Создаем SHA1 хеш из входного текста
	hash := sha1.New()
	hash.Write([]byte(r.Text))
	hashBytes := hash.Sum(nil)
	hashString := hex.EncodeToString(hashBytes)

	// 2. Определяем какие части использовать на основе хеша
	parts := make(map[string]string)

	switch r.Set {
	case "set1":
		colorParts, err := filepath.Glob(filepath.Join(assetsDir, r.Set, "*"))
		if err != nil || len(colorParts) == 0 {
			return nil, err
		}
		colorIndex := hexToInt(hashString[0:2]) % len(colorParts)
		color := filepath.Base(colorParts[colorIndex])

		parts["body"] = selectPart(hashString[2:4], filepath.Join(r.Set, color, "003#01Body"))
		parts["face"] = selectPart(hashString[4:6], filepath.Join(r.Set, color, "004#02Face"))
		parts["eyes"] = selectPart(hashString[6:8], filepath.Join(r.Set, color, "001#Eyes"))
		parts["mouth"] = selectPart(hashString[8:10], filepath.Join(r.Set, color, "000#Mouth"))
		parts["accessory"] = selectPart(hashString[10:12], filepath.Join(r.Set, color, "002#Accessory"))

	case "set2":
		parts["body"] = selectPart(hashString[0:2], filepath.Join(r.Set, "000#04Body"))
		parts["mouth"] = selectPart(hashString[2:4], filepath.Join(r.Set, "001#Mouth"))
		parts["eyes"] = selectPart(hashString[4:6], filepath.Join(r.Set, "002#Eyes"))
		parts["face"] = selectPart(hashString[6:8], filepath.Join(r.Set, "006#03Faces"))

	case "set3":
		parts["base"] = selectPart(hashString[0:4], filepath.Join(r.Set, "005#01BaseFace"))

		parts["eyes"] = selectPart(hashString[4:6], filepath.Join(r.Set, "003#04Eyes"))

		if eyebrows := selectPart(hashString[6:8], filepath.Join(r.Set, "002#05Eyebrows")); eyebrows != "" {
			parts["eyebrows"] = eyebrows
		}

		if nose := selectPart(hashString[8:10], filepath.Join(r.Set, "004#06Nose")); nose != "" {
			parts["nose"] = nose
		}

		parts["mouth"] = selectPart(hashString[10:12], filepath.Join(r.Set, "000#07Mouth"))

		if antenna := selectPart(hashString[12:14], filepath.Join(r.Set, "006#03Antenna")); antenna != "" {
			parts["antenna"] = antenna
		}

	case "set4":
		parts["body"] = selectPart(hashString[0:2], filepath.Join(r.Set, "000#00body"))
		parts["fur"] = selectPart(hashString[2:4], filepath.Join(r.Set, "001#01fur"))
		parts["eyes"] = selectPart(hashString[4:6], filepath.Join(r.Set, "002#02eyes"))
		parts["mouth"] = selectPart(hashString[6:8], filepath.Join(r.Set, "003#03mouth"))
		parts["accessory"] = selectPart(hashString[8:10], filepath.Join(r.Set, "004#04accessories"))

	case "set5":
		parts["body"] = selectPart(hashString[0:2], filepath.Join(r.Set, "000#Body"))
		parts["eyes"] = selectPart(hashString[2:4], filepath.Join(r.Set, "001#Eye"))
		parts["eyebrow"] = selectPart(hashString[4:6], filepath.Join(r.Set, "002#Eyebrow"))
		parts["mouth"] = selectPart(hashString[6:8], filepath.Join(r.Set, "003#Mouth"))
		parts["clothes"] = selectPart(hashString[8:10], filepath.Join(r.Set, "004#Cloth"))

	default:
		return nil, fmt.Errorf("unknown set: %s", r.Set)
	}

	return composeImage(parts, r.Size, r.BGSet, r.Set, r.Text)
}

func hexToInt(hexStr string) int {
	num, err := strconv.ParseInt(hexStr, 16, 64)
	if err != nil {
		return 0
	}
	return int(num)
}

func selectPart(hashPart string, partPath string) string {
	pattern := filepath.Join(assetsDir, partPath, "*.png")
	files, err := filepath.Glob(pattern)
	if err != nil || len(files) == 0 {
		log.Printf("No files found for pattern: %s", pattern)
		return ""
	}

	index := hexToInt(hashPart) % len(files)
	return files[index]
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

func composeImage(parts map[string]string, size string, bgSet string, set string, text string) (image.Image, error) {
	// Получаем размеры для текущего набора
	width, height := getSetDimensions(set)
	base := image.NewRGBA(image.Rect(0, 0, width, height))

	// Выбираем и загружаем фон если указан bgSet
	if bgSet != "" {
		bgPattern := filepath.Join(assetsDir, "backgrounds", bgSet, "*.png")
		bgFiles, err := filepath.Glob(bgPattern)
		if err != nil {
			return nil, fmt.Errorf("error finding background files: %v", err)
		}

		if len(bgFiles) > 0 {
			// Генерируем индекс фона на основе хеша текста
			hash := sha1.New()
			hash.Write([]byte(text))
			hashBytes := hash.Sum(nil)
			hashInt := binary.BigEndian.Uint64(hashBytes)
			bgIndex := int(hashInt % uint64(len(bgFiles)))

			// Загружаем и масштабируем выбранный фон
			bgImg, err := loadAndResizeImage(bgFiles[bgIndex], width, height)
			if err != nil {
				return nil, fmt.Errorf("error loading background: %v", err)
			}
			draw.Draw(base, base.Bounds(), bgImg, image.Point{}, draw.Over)
		}
	}

	// Получаем порядок отрисовки
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

	// Дополнительное масштабирование если указан параметр size
	if size != "" {
		return resizeImage(base, size)
	}
	if base != nil {
		return base, nil
	} else {
		return nil, fmt.Errorf("failed to generate image")
	}

}

// getPartsOrder возвращает порядок отрисовки частей для каждого набора
func getPartsOrder(set string) []string {
	switch set {
	case "set1":
		// Порядок для классических роботов (цветные наборы)
		return []string{
			"body",      // 003#01Body
			"face",      // 004#02Face
			"eyes",      // 001#Eyes
			"mouth",     // 000#Mouth
			"accessory", // 002#Accessory
		}

	case "set2":
		// Порядок для монстров
		return []string{
			"body",       // 000#04Body
			"face",       // 006#03Faces
			"mouth",      // 001#Mouth
			"eyes",       // 002#Eyes
			"bodycolors", // 003#02BodyColors
			"facecolors", // 004#01FaceColors
			"nose",       // 005#Nose
		}

	case "set3":
		// Порядок для голов роботов (самый сложный набор)
		return []string{
			"base",     // 005#01BaseFace - основной слой лица
			"wave",     // 001#02Wave - волны/фон (если есть)
			"eyes",     // 003#04Eyes - глаза
			"eyebrows", // 002#05Eyebrows - брови
			"nose",     // 004#06Nose - нос
			"mouth",    // 000#07Mouth - рот
			"antenna",  // 006#03Antenna - антенна
		}

	case "set4":
		// Порядок для котов
		return []string{
			"body",      // 000#00body
			"fur",       // 001#01fur
			"eyes",      // 002#02eyes
			"mouth",     // 003#03mouth
			"accessory", // 004#04accessories
		}

	case "set5":
		// Порядок для человеческих аватаров
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

	// Если изображение уже нужного размера, возвращаем как есть
	if img.Bounds().Dx() == width && img.Bounds().Dy() == height {
		return img, nil
	}

	// Масштабируем изображение
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
	file, err := os.Open(path)
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
