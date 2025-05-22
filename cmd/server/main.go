package main

import (
	"fmt"
	"image/png"
	"log"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/terem42/robohash/robohash"
)

var buildVersion = "HEAD"

func hashHandler(w http.ResponseWriter, r *http.Request) {
	// Извлекаем текст из URL (формат /text.png)
	path := strings.TrimPrefix(r.URL.Path, "/")
	text := strings.TrimSuffix(path, filepath.Ext(path))

	if text == "" {
		text = "example" // Значение по умолчанию
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

	// Отправляем изображение
	w.Header().Set("Content-Type", "image/png")
	if err := png.Encode(w, img); err != nil {
		http.Error(w, fmt.Sprintf("Error encoding image: %v", err), http.StatusInternalServerError)
		return
	}
}

func main() {
	log.Printf("Robohash Go version %s", buildVersion)
	http.HandleFunc("/", hashHandler)
	http.Handle("/assets/", http.StripPrefix("/assets/", http.FileServer(http.Dir("assets"))))

	fmt.Println("Server running on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
