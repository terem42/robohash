package main

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"log"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/davidbyttow/govips/v2/vips"
	"github.com/terem42/robohash/robohash"
)

var buildVersion = "HEAD"

func generateETag(data []byte) string {
	hash := sha256.Sum256(data)
	return hex.EncodeToString(hash[:])
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status": "ok", "version": "` + buildVersion + `", "timestamp": "` + time.Now().UTC().Format(time.RFC3339) + `"}`))
}

func hashHandler(w http.ResponseWriter, r *http.Request) {

	path := strings.TrimPrefix(r.URL.Path, "/")
	ext := filepath.Ext(path)
	text := strings.TrimSuffix(path, filepath.Ext(path))

	if strings.HasPrefix(path, "favicon") {
		http.NotFound(w, r)
		return
	}

	if text == "" {
		text = "example"
	}

	query := r.URL.Query()
	roboHash := robohash.RoboHash{
		Text:  text,
		Set:   query.Get("set"),
		Size:  query.Get("size"),
		BGSet: query.Get("bgset"),
	}

	img, err := roboHash.Generate()
	if err != nil {
		http.Error(w, fmt.Sprintf("Error generating image: %v", err), http.StatusInternalServerError)
		return
	}
	defer img.Close()

	var imgBuf []byte
	var contentType string

	switch strings.ToLower(ext) {
	case ".avif":
		// Экспорт в AVIF
		imgBuf, _, err = img.ExportAvif(&vips.AvifExportParams{
			Quality:  85,    // Качество сжатия
			Speed:    8,     // Скорость кодирования (0-8, больше = быстрее но хуже качество)
			Lossless: false, // Сжатие с потерями
		})
		if err != nil {
			http.Error(w, fmt.Sprintf("Error encoding AVIF image: %v", err), http.StatusInternalServerError)
			return
		}
		contentType = "image/avif"

	case ".webp":
		// Экспорт в WebP
		imgBuf, _, err = img.ExportWebp(&vips.WebpExportParams{
			Quality:         85,   // Качество для lossy
			Lossless:        true, // Используем lossless для лучшего качества
			NearLossless:    false,
			ReductionEffort: 4, // Уровень оптимизации (0-6)
		})
		if err != nil {
			http.Error(w, fmt.Sprintf("Error encoding WEBP image: %v", err), http.StatusInternalServerError)
			return
		}
		contentType = "image/webp"

	case ".jpg", ".jpeg":
		// Экспорт в JPEG
		imgBuf, _, err = img.ExportJpeg(&vips.JpegExportParams{
			Quality:        85,
			Interlace:      false,
			OptimizeCoding: true,
			SubsampleMode:  vips.VipsForeignSubsampleAuto,
		})
		if err != nil {
			http.Error(w, fmt.Sprintf("Error encoding JPEG image: %v", err), http.StatusInternalServerError)
			return
		}
		contentType = "image/jpeg"

	default:
		// Экспорт в PNG (по умолчанию)
		imgBuf, _, err = img.ExportPng(&vips.PngExportParams{
			Compression: 6,     // Уровень сжатия PNG (0-9)
			Interlace:   false, // Прогрессивная загрузка
			Quality:     85,    // Качество (для палитровых изображений)
		})
		if err != nil {
			http.Error(w, fmt.Sprintf("Error encoding PNG image: %v", err), http.StatusInternalServerError)
			return
		}
		contentType = "image/png"
	}

	// Устанавливаем заголовки ответа
	w.Header().Set("Content-Type", contentType)
	w.Header().Set("Cache-Control", "public, max-age=31536000")
	w.Header().Set("Content-Length", strconv.Itoa(len(imgBuf)))
	w.Header().Set("ETag", `"`+generateETag(imgBuf)+`"`)
	w.Header().Set("Last-Modified", time.Now().UTC().Format(http.TimeFormat))
	w.Write(imgBuf)

}

func main() {
	log.Printf("Robohash Go version %s", buildVersion)
	http.HandleFunc("/health", healthHandler)
	http.HandleFunc("/", hashHandler)
	fmt.Println("Server running on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
