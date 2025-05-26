package robohash

import (
	"container/list"
	"image"
	"os"
	"strconv"
	"sync"
)

type ImageCache struct {
	maxSize     int
	currentSize int
	mu          sync.Mutex
	list        *list.List
	items       map[string]*list.Element
}

type cacheItem struct {
	key  string
	img  image.Image
	size int
}

func getCacheSize() int {
	defaultSize := 50

	if sizeStr := os.Getenv("ROBOHASH_IMG_CACHE_SIZE"); sizeStr != "" {
		if size, err := strconv.Atoi(sizeStr); err == nil && size > 0 {
			return size
		}
	}
	return defaultSize
}

func NewImageCache() *ImageCache {
	maxSizeMB := getCacheSize()
	return &ImageCache{
		maxSize: maxSizeMB * 1024 * 1024,
		list:    list.New(),
		items:   make(map[string]*list.Element),
	}
}

func (c *ImageCache) Get(key string) (image.Image, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if elem, ok := c.items[key]; ok {
		c.list.MoveToFront(elem)
		return elem.Value.(*cacheItem).img, true
	}
	return nil, false
}

func (c *ImageCache) Set(key string, img image.Image, imgSize int) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if elem, ok := c.items[key]; ok {
		c.list.MoveToFront(elem)
		oldSize := elem.Value.(*cacheItem).size
		elem.Value.(*cacheItem).img = img
		elem.Value.(*cacheItem).size = imgSize
		c.currentSize += imgSize - oldSize
		return
	}

	for c.currentSize+imgSize > c.maxSize && c.list.Len() > 0 {
		oldest := c.list.Back()
		c.list.Remove(oldest)
		delete(c.items, oldest.Value.(*cacheItem).key)
		c.currentSize -= oldest.Value.(*cacheItem).size
	}

	item := &cacheItem{key: key, img: img, size: imgSize}
	elem := c.list.PushFront(item)
	c.items[key] = elem
	c.currentSize += imgSize
}
