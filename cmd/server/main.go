package main

import (
	"fmt"
	"image/png"
	"log"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/Kagami/go-avif"
	"github.com/terem42/robohash/robohash"
)

var buildVersion = "HEAD"

func hashHandler(w http.ResponseWriter, r *http.Request) {
	// Извлекаем текст из URL (формат /text.png)
	path := strings.TrimPrefix(r.URL.Path, "/")
	ext := filepath.Ext(path)
	text := strings.TrimSuffix(path, filepath.Ext(path))

	if strings.HasPrefix(r.URL.Path, "favicon") {
		http.NotFound(w, r)
	}

	if text == "" {
		text = "example"
	}

	// Парсим параметры запроса
	query := r.URL.Query()
	roboHash := robohash.RoboHash{
		Text:  text,
		Set:   query.Get("set"),
		Size:  query.Get("size"),
		BGSet: query.Get("bgset"),
	}

	// Генерируем изображение
	img, err := roboHash.Generate()
	if err != nil {
		http.Error(w, fmt.Sprintf("Error generating image: %v", err), http.StatusInternalServerError)
		return
	}

	// Определяем формат по расширению файла
	switch strings.ToLower(ext) {
	case ".avif":
		// Кодируем в AVIF
		w.Header().Set("Content-Type", "image/avif")
		if err := avif.Encode(w, img, nil); err != nil {
			http.Error(w, fmt.Sprintf("Error encoding AVIF image: %v", err), http.StatusInternalServerError)
			return
		}
	default:
		// По умолчанию кодируем в PNG
		w.Header().Set("Content-Type", "image/png")
		if err := png.Encode(w, img); err != nil {
			http.Error(w, fmt.Sprintf("Error encoding PNG image: %v", err), http.StatusInternalServerError)
			return
		}
	}

}

func main() {
	log.Printf("Robohash Go version %s", buildVersion)
	http.HandleFunc("/", hashHandler)
	fmt.Println("Server running on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
