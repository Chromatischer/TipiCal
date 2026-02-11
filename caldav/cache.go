package caldav

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// CacheEntry represents a cached calendar object.
type CacheEntry struct {
	Path      string    `json:"path"`
	ETag      string    `json:"etag"`
	Data      string    `json:"data"` // raw iCal data
	FetchedAt time.Time `json:"fetched_at"`
}

// CacheIndex represents the cache metadata for a calendar.
type CacheIndex struct {
	CalendarURL   string       `json:"calendar_url"`
	SyncToken     string       `json:"sync_token"`
	LastSync      time.Time    `json:"last_sync"`
	CalendarName  string       `json:"calendar_name"`
	CalendarColor string       `json:"calendar_color"`
	ConfigIndex   int          `json:"config_index"`
	Entries       []CacheEntry `json:"entries"`
}

// Cache manages local caching of CalDAV objects.
type Cache struct {
	dir   string
	mu    sync.RWMutex
	index map[string]*CacheIndex // key: calendar URL
}

// NewCache creates a cache in the specified directory.
func NewCache(cacheDir string) (*Cache, error) {
	if err := os.MkdirAll(cacheDir, 0700); err != nil {
		return nil, err
	}
	cache := &Cache{
		dir:   cacheDir,
		index: make(map[string]*CacheIndex),
	}
	cache.loadIndex()
	return cache, nil
}

// Get retrieves a cached calendar entry by path.
func (c *Cache) Get(calURL, path string) *CacheEntry {
	c.mu.RLock()
	defer c.mu.RUnlock()

	idx, ok := c.index[calURL]
	if !ok {
		return nil
	}

	for _, entry := range idx.Entries {
		if entry.Path == path {
			return &entry
		}
	}
	return nil
}

// GetAll retrieves all cached entries for a calendar.
func (c *Cache) GetAll(calURL string) []CacheEntry {
	c.mu.RLock()
	defer c.mu.RUnlock()

	idx, ok := c.index[calURL]
	if !ok {
		return nil
	}
	return idx.Entries
}

// AllIndexes returns all cached calendar indexes.
func (c *Cache) AllIndexes() map[string]*CacheIndex {
	c.mu.RLock()
	defer c.mu.RUnlock()

	result := make(map[string]*CacheIndex, len(c.index))
	for k, v := range c.index {
		result[k] = v
	}
	return result
}

// SetIndexMeta updates the calendar metadata on a cache index.
func (c *Cache) SetIndexMeta(calURL, name, color string, configIndex int) {
	c.mu.Lock()
	defer c.mu.Unlock()

	idx, ok := c.index[calURL]
	if !ok {
		idx = &CacheIndex{CalendarURL: calURL}
		c.index[calURL] = idx
	}
	idx.CalendarName = name
	idx.CalendarColor = color
	idx.ConfigIndex = configIndex
	c.saveIndex()
}

// Put stores or updates a cached calendar entry.
func (c *Cache) Put(calURL string, entry CacheEntry) {
	c.mu.Lock()
	defer c.mu.Unlock()

	idx, ok := c.index[calURL]
	if !ok {
		idx = &CacheIndex{
			CalendarURL: calURL,
		}
		c.index[calURL] = idx
	}

	// Update or append
	found := false
	for i, e := range idx.Entries {
		if e.Path == entry.Path {
			idx.Entries[i] = entry
			found = true
			break
		}
	}
	if !found {
		idx.Entries = append(idx.Entries, entry)
	}

	idx.LastSync = time.Now()
	c.saveIndex()
}

// Remove removes a cached entry.
func (c *Cache) Remove(calURL, path string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	idx, ok := c.index[calURL]
	if !ok {
		return
	}

	for i, e := range idx.Entries {
		if e.Path == path {
			idx.Entries = append(idx.Entries[:i], idx.Entries[i+1:]...)
			break
		}
	}
	c.saveIndex()
}

// LastSyncTime returns the last sync time for a calendar.
func (c *Cache) LastSyncTime(calURL string) time.Time {
	c.mu.RLock()
	defer c.mu.RUnlock()

	idx, ok := c.index[calURL]
	if !ok {
		return time.Time{}
	}
	return idx.LastSync
}

func (c *Cache) indexPath() string {
	return filepath.Join(c.dir, "cache_index.json")
}

func (c *Cache) loadIndex() {
	data, err := os.ReadFile(c.indexPath())
	if err != nil {
		return
	}

	var indexes map[string]*CacheIndex
	if err := json.Unmarshal(data, &indexes); err != nil {
		return
	}
	c.index = indexes
}

func (c *Cache) saveIndex() {
	data, err := json.MarshalIndent(c.index, "", "  ")
	if err != nil {
		return
	}
	_ = os.WriteFile(c.indexPath(), data, 0600)
}
