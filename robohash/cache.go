package robohash

import (
	"container/list"
	"image"
	"os"
	"strconv"
	"sync"

	"github.com/cespare/xxhash/v2"
)

const (
	shardSizeMB   = 10
	minShardCount = 4
)

type cacheItem struct {
	key  string
	img  image.Image
	size int
}

type ShardedImageCache struct {
	shards []*imageCacheShard
}

type imageCacheShard struct {
	mu          sync.RWMutex
	list        *list.List
	items       map[string]*list.Element
	currentSize int
	maxSize     int
}

func NewImageCache() *ShardedImageCache {
	maxSizeMB := getCacheSizeMB()
	shardCount := calculateShardCount(maxSizeMB)

	shards := make([]*imageCacheShard, shardCount)
	for i := range shards {
		shards[i] = &imageCacheShard{
			list:    list.New(),
			items:   make(map[string]*list.Element),
			maxSize: (maxSizeMB * 1024 * 1024) / shardCount,
		}
	}

	return &ShardedImageCache{
		shards: shards,
	}
}

func calculateShardCount(maxSizeMB int) int {
	shardCount := maxSizeMB / shardSizeMB
	if shardCount < minShardCount {
		return minShardCount
	}
	return shardCount
}

func getCacheSizeMB() int {
	if sizeStr := os.Getenv("ROBOHASH_IMG_CACHE_SIZE"); sizeStr != "" {
		if size, err := strconv.Atoi(sizeStr); err == nil && size > 0 {
			return size
		}
	}
	return 100 // Default size 100MB
}

func (c *ShardedImageCache) getShard(key string) *imageCacheShard {
	hash := xxhash.Sum64String(key)
	return c.shards[hash%uint64(len(c.shards))]
}

func (c *ShardedImageCache) Get(key string) (image.Image, bool) {
	shard := c.getShard(key)
	shard.mu.RLock()
	defer shard.mu.RUnlock()

	if elem, ok := shard.items[key]; ok {
		shard.list.MoveToFront(elem)
		return elem.Value.(*cacheItem).img, true
	}
	return nil, false
}

func (c *ShardedImageCache) Set(key string, img image.Image, imgSize int) {
	shard := c.getShard(key)
	shard.mu.Lock()
	defer shard.mu.Unlock()

	if elem, ok := shard.items[key]; ok {
		oldItem := elem.Value.(*cacheItem)
		oldSize := oldItem.size
		oldItem.img = img
		oldItem.size = imgSize
		shard.currentSize += imgSize - oldSize
		return
	}

	for shard.currentSize+imgSize > shard.maxSize && shard.list.Len() > 0 {
		oldest := shard.list.Back()
		shard.list.Remove(oldest)
		delete(shard.items, oldest.Value.(*cacheItem).key)
		shard.currentSize -= oldest.Value.(*cacheItem).size
	}

	item := &cacheItem{
		key:  key,
		img:  img,
		size: imgSize,
	}
	elem := shard.list.PushFront(item)
	shard.items[key] = elem
	shard.currentSize += imgSize
}
