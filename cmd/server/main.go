package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"image/png"
	"log"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/Kagami/go-avif"
	"github.com/chai2010/webp"
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

	imgBuf := new(bytes.Buffer)

	switch strings.ToLower(ext) {
	case ".avif":
		if err := avif.Encode(imgBuf, img, nil); err != nil {
			http.Error(w, fmt.Sprintf("Error encoding AVIF image: %v", err), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "image/avif")
	case ".webp":
		if err := webp.Encode(imgBuf, img, &webp.Options{Lossless: true}); err != nil {
			http.Error(w, fmt.Sprintf("Error encoding WEBP image: %v", err), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "image/webp")
	default:
		if err := png.Encode(imgBuf, img); err != nil {
			http.Error(w, fmt.Sprintf("Error encoding PNG image: %v", err), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "image/png")
	}
	w.Header().Set("Cache-Control", "public, max-age=31536000")
	w.Header().Set("Content-Length", strconv.Itoa(len(imgBuf.Bytes())))
	w.Header().Set("ETag", `"`+generateETag(imgBuf.Bytes())+`"`)
	w.Header().Set("Last-Modified", time.Now().UTC().Format(http.TimeFormat))
	w.Write(imgBuf.Bytes())

}

func main() {
	log.Printf("Robohash Go version %s", buildVersion)
	http.HandleFunc("/health", healthHandler)
	http.HandleFunc("/", hashHandler)
	fmt.Println("Server running on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
