package robohash

import (
	"container/list"
	"image"
	"sync"
)

type ImageCache struct {
	maxSize     int
	currentSize int
	mu          sync.Mutex
	list        *list.List               // Для отслеживания порядка использования
	items       map[string]*list.Element // Для быстрого доступа
}

type cacheItem struct {
	key  string
	img  image.Image
	size int // Примерный размер в байтах
}

func NewImageCache(maxSizeMB int) *ImageCache {
	return &ImageCache{
		maxSize: maxSizeMB * 1024 * 1024, // Конвертируем МБ в байты
		list:    list.New(),
		items:   make(map[string]*list.Element),
	}
}

func (c *ImageCache) Get(key string) (image.Image, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if elem, ok := c.items[key]; ok {
		c.list.MoveToFront(elem) // Обновляем как "использованный"
		return elem.Value.(*cacheItem).img, true
	}
	return nil, false
}

func (c *ImageCache) Set(key string, img image.Image, imgSize int) {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Если элемент уже есть - обновляем
	if elem, ok := c.items[key]; ok {
		c.list.MoveToFront(elem)
		oldSize := elem.Value.(*cacheItem).size
		elem.Value.(*cacheItem).img = img
		elem.Value.(*cacheItem).size = imgSize
		c.currentSize += imgSize - oldSize
		return
	}

	// Вытесняем старые элементы, если не хватает места
	for c.currentSize+imgSize > c.maxSize && c.list.Len() > 0 {
		oldest := c.list.Back()
		c.list.Remove(oldest)
		delete(c.items, oldest.Value.(*cacheItem).key)
		c.currentSize -= oldest.Value.(*cacheItem).size
	}

	// Добавляем новый элемент
	item := &cacheItem{key: key, img: img, size: imgSize}
	elem := c.list.PushFront(item)
	c.items[key] = elem
	c.currentSize += imgSize
}
