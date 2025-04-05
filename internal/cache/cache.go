package cache

import (
	"sync"
	"time"
)

// CacheItem представляет элемент кэша
type CacheItem struct {
	Value      interface{}
	Expiration int64
}

// Cache представляет структуру кэша
type Cache struct {
	items map[string]CacheItem
	mu    sync.RWMutex
}

// NewCache создает новый экземпляр кэша
func NewCache() *Cache {
	cache := &Cache{
		items: make(map[string]CacheItem),
	}
	go cache.cleanupLoop() // запуск очистки кэша
	return cache
}

// Set добавляет значение в кэш
func (c *Cache) Set(key string, value interface{}, duration time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()

	// установка времени истечения кэша
	var expiration int64
	if duration > 0 {
		expiration = time.Now().Add(duration).UnixNano()
	}

	// добавление значения в кэш
	c.items[key] = CacheItem{
		Value:      value,
		Expiration: expiration,
	}
}

// Get получает значение из кэша
func (c *Cache) Get(key string) (interface{}, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	// получение значения из кэша
	item, exists := c.items[key]
	if !exists {
		return nil, false
	}

	// проверка на истечение времени
	if item.Expiration > 0 && time.Now().UnixNano() > item.Expiration {
		return nil, false
	}

	return item.Value, true // возвращает значение и true, если оно существует и не истекло
}

// Delete удаляет значение из кэша
func (c *Cache) Delete(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	delete(c.items, key)
}

// cleanupLoop очищает устаревшие элементы кэша
func (c *Cache) cleanupLoop() {
	ticker := time.NewTicker(time.Minute) // установка интервала очистки кэша
	for range ticker.C {
		c.mu.Lock()
		now := time.Now().UnixNano()
		for key, item := range c.items { // проверка на истечение времени
			if item.Expiration > 0 && now > item.Expiration {
				delete(c.items, key)
			}
		}
		c.mu.Unlock()
	}
}
