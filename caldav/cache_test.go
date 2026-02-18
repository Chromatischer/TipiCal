package caldav

import (
	"os"
	"testing"
	"time"
)

func tempCacheDir(t *testing.T) string {
	dir, err := os.MkdirTemp("", "tipical-cache-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	t.Cleanup(func() { os.RemoveAll(dir) })
	return dir
}

func TestCacheNew(t *testing.T) {
	dir := tempCacheDir(t)
	cache, err := NewCache(dir)
	if err != nil {
		t.Fatalf("NewCache() error = %v", err)
	}
	if cache == nil {
		t.Fatal("NewCache() returned nil")
	}
	if cache.dir != dir {
		t.Errorf("cache.dir = %q, want %q", cache.dir, dir)
	}
}

func TestCachePutAndGet(t *testing.T) {
	dir := tempCacheDir(t)
	cache, err := NewCache(dir)
	if err != nil {
		t.Fatalf("NewCache() error = %v", err)
	}

	calURL := "https://example.com/calendars/work/"
	entry := CacheEntry{
		Path:      "/calendars/work/event1.ics",
		ETag:      "etag-123",
		Data:      "BEGIN:VCALENDAR...",
		FetchedAt: time.Now(),
	}

	cache.Put(calURL, entry)

	got := cache.Get(calURL, entry.Path)
	if got == nil {
		t.Fatal("Get() returned nil")
	}
	if got.Path != entry.Path {
		t.Errorf("Path = %q, want %q", got.Path, entry.Path)
	}
	if got.ETag != entry.ETag {
		t.Errorf("ETag = %q, want %q", got.ETag, entry.ETag)
	}
	if got.Data != entry.Data {
		t.Errorf("Data = %q, want %q", got.Data, entry.Data)
	}
}

func TestCacheGetNonExistent(t *testing.T) {
	dir := tempCacheDir(t)
	cache, err := NewCache(dir)
	if err != nil {
		t.Fatalf("NewCache() error = %v", err)
	}

	got := cache.Get("https://nonexistent.com/", "/path.ics")
	if got != nil {
		t.Errorf("Get() for non-existent entry = %v, want nil", got)
	}
}

func TestCacheGetAll(t *testing.T) {
	dir := tempCacheDir(t)
	cache, err := NewCache(dir)
	if err != nil {
		t.Fatalf("NewCache() error = %v", err)
	}

	calURL := "https://example.com/calendars/work/"

	for i := 0; i < 3; i++ {
		cache.Put(calURL, CacheEntry{
			Path:      string(rune('a' + i)),
			ETag:      "etag",
			Data:      "data",
			FetchedAt: time.Now(),
		})
	}

	entries := cache.GetAll(calURL)
	if len(entries) != 3 {
		t.Errorf("GetAll() returned %d entries, want 3", len(entries))
	}
}

func TestCacheGetAllEmpty(t *testing.T) {
	dir := tempCacheDir(t)
	cache, err := NewCache(dir)
	if err != nil {
		t.Fatalf("NewCache() error = %v", err)
	}

	entries := cache.GetAll("https://nonexistent.com/")
	if entries != nil {
		t.Errorf("GetAll() for non-existent calendar = %v, want nil", entries)
	}
}

func TestCacheUpdateExisting(t *testing.T) {
	dir := tempCacheDir(t)
	cache, err := NewCache(dir)
	if err != nil {
		t.Fatalf("NewCache() error = %v", err)
	}

	calURL := "https://example.com/cal/"
	entry := CacheEntry{
		Path:      "/event.ics",
		ETag:      "etag-1",
		Data:      "data-1",
		FetchedAt: time.Now(),
	}
	cache.Put(calURL, entry)

	updated := CacheEntry{
		Path:      "/event.ics",
		ETag:      "etag-2",
		Data:      "data-2",
		FetchedAt: time.Now(),
	}
	cache.Put(calURL, updated)

	got := cache.Get(calURL, entry.Path)
	if got == nil {
		t.Fatal("Get() returned nil")
	}
	if got.ETag != "etag-2" {
		t.Errorf("ETag = %q, want %q", got.ETag, "etag-2")
	}
	if got.Data != "data-2" {
		t.Errorf("Data = %q, want %q", got.Data, "data-2")
	}

	all := cache.GetAll(calURL)
	if len(all) != 1 {
		t.Errorf("GetAll() returned %d entries, want 1 (should update, not add)", len(all))
	}
}

func TestCacheRemove(t *testing.T) {
	dir := tempCacheDir(t)
	cache, err := NewCache(dir)
	if err != nil {
		t.Fatalf("NewCache() error = %v", err)
	}

	calURL := "https://example.com/cal/"
	entry := CacheEntry{
		Path:      "/event.ics",
		ETag:      "etag",
		Data:      "data",
		FetchedAt: time.Now(),
	}
	cache.Put(calURL, entry)

	cache.Remove(calURL, entry.Path)

	got := cache.Get(calURL, entry.Path)
	if got != nil {
		t.Error("Get() should return nil after Remove()")
	}
}

func TestCacheRemoveNonExistent(t *testing.T) {
	dir := tempCacheDir(t)
	cache, err := NewCache(dir)
	if err != nil {
		t.Fatalf("NewCache() error = %v", err)
	}

	cache.Remove("https://nonexistent.com/", "/path.ics")
}

func TestCacheSetIndexMeta(t *testing.T) {
	dir := tempCacheDir(t)
	cache, err := NewCache(dir)
	if err != nil {
		t.Fatalf("NewCache() error = %v", err)
	}

	calURL := "https://example.com/cal/"
	cache.SetIndexMeta(calURL, "Work Calendar", "#3B82F6", "Work Account", 0)

	indexes := cache.AllIndexes()
	idx, ok := indexes[calURL]
	if !ok {
		t.Fatal("index not found")
	}
	if idx.CalendarName != "Work Calendar" {
		t.Errorf("CalendarName = %q, want %q", idx.CalendarName, "Work Calendar")
	}
	if idx.CalendarColor != "#3B82F6" {
		t.Errorf("CalendarColor = %q, want %q", idx.CalendarColor, "#3B82F6")
	}
	if idx.SourceName != "Work Account" {
		t.Errorf("SourceName = %q, want %q", idx.SourceName, "Work Account")
	}
	if idx.ConfigIndex != 0 {
		t.Errorf("ConfigIndex = %d, want 0", idx.ConfigIndex)
	}
}

func TestCacheAllIndexes(t *testing.T) {
	dir := tempCacheDir(t)
	cache, err := NewCache(dir)
	if err != nil {
		t.Fatalf("NewCache() error = %v", err)
	}

	calURL1 := "https://example.com/cal1/"
	calURL2 := "https://example.com/cal2/"

	cache.Put(calURL1, CacheEntry{Path: "/event1.ics"})
	cache.Put(calURL2, CacheEntry{Path: "/event2.ics"})

	indexes := cache.AllIndexes()
	if len(indexes) != 2 {
		t.Errorf("AllIndexes() returned %d indexes, want 2", len(indexes))
	}
}

func TestCacheAllIndexesEmpty(t *testing.T) {
	dir := tempCacheDir(t)
	cache, err := NewCache(dir)
	if err != nil {
		t.Fatalf("NewCache() error = %v", err)
	}

	indexes := cache.AllIndexes()
	if len(indexes) != 0 {
		t.Errorf("AllIndexes() on empty cache returned %d indexes, want 0", len(indexes))
	}
}

func TestCacheLastSyncTime(t *testing.T) {
	dir := tempCacheDir(t)
	cache, err := NewCache(dir)
	if err != nil {
		t.Fatalf("NewCache() error = %v", err)
	}

	calURL := "https://example.com/cal/"

	before := cache.LastSyncTime(calURL)
	if !before.IsZero() {
		t.Errorf("LastSyncTime for non-existent calendar = %v, want zero", before)
	}

	cache.Put(calURL, CacheEntry{Path: "/event.ics"})

	after := cache.LastSyncTime(calURL)
	if after.IsZero() {
		t.Error("LastSyncTime should not be zero after Put()")
	}

	time.Sleep(10 * time.Millisecond)
	cache.Put(calURL, CacheEntry{Path: "/event2.ics"})
	later := cache.LastSyncTime(calURL)
	if later.Before(after) {
		t.Error("LastSyncTime should be updated after subsequent Put()")
	}
}

func TestCachePersistence(t *testing.T) {
	dir := tempCacheDir(t)

	cache1, err := NewCache(dir)
	if err != nil {
		t.Fatalf("NewCache() error = %v", err)
	}

	calURL := "https://example.com/cal/"
	entry := CacheEntry{
		Path:      "/event.ics",
		ETag:      "etag-123",
		Data:      "BEGIN:VCALENDAR...",
		FetchedAt: time.Date(2026, time.February, 18, 12, 0, 0, 0, time.UTC),
	}
	cache1.Put(calURL, entry)
	cache1.SetIndexMeta(calURL, "Work", "#3B82F6", "Account", 0)

	cache2, err := NewCache(dir)
	if err != nil {
		t.Fatalf("NewCache() second instance error = %v", err)
	}

	got := cache2.Get(calURL, entry.Path)
	if got == nil {
		t.Fatal("Get() after reload returned nil")
	}
	if got.Path != entry.Path {
		t.Errorf("Path after reload = %q, want %q", got.Path, entry.Path)
	}
	if got.ETag != entry.ETag {
		t.Errorf("ETag after reload = %q, want %q", got.ETag, entry.ETag)
	}

	indexes := cache2.AllIndexes()
	idx, ok := indexes[calURL]
	if !ok {
		t.Fatal("index not found after reload")
	}
	if idx.CalendarName != "Work" {
		t.Errorf("CalendarName after reload = %q, want %q", idx.CalendarName, "Work")
	}
}

func TestCacheMultipleCalendars(t *testing.T) {
	dir := tempCacheDir(t)
	cache, err := NewCache(dir)
	if err != nil {
		t.Fatalf("NewCache() error = %v", err)
	}

	calURL1 := "https://example.com/work/"
	calURL2 := "https://example.com/personal/"

	cache.Put(calURL1, CacheEntry{Path: "/work1.ics", Data: "work-data-1"})
	cache.Put(calURL1, CacheEntry{Path: "/work2.ics", Data: "work-data-2"})
	cache.Put(calURL2, CacheEntry{Path: "/personal1.ics", Data: "personal-data-1"})

	entries1 := cache.GetAll(calURL1)
	if len(entries1) != 2 {
		t.Errorf("GetAll(work) = %d entries, want 2", len(entries1))
	}

	entries2 := cache.GetAll(calURL2)
	if len(entries2) != 1 {
		t.Errorf("GetAll(personal) = %d entries, want 1", len(entries2))
	}

	indexes := cache.AllIndexes()
	if len(indexes) != 2 {
		t.Errorf("AllIndexes() = %d indexes, want 2", len(indexes))
	}
}

func TestCacheRemoveFromSpecificCalendar(t *testing.T) {
	dir := tempCacheDir(t)
	cache, err := NewCache(dir)
	if err != nil {
		t.Fatalf("NewCache() error = %v", err)
	}

	calURL1 := "https://example.com/work/"
	calURL2 := "https://example.com/personal/"

	cache.Put(calURL1, CacheEntry{Path: "/shared.ics", Data: "work-data"})
	cache.Put(calURL2, CacheEntry{Path: "/shared.ics", Data: "personal-data"})

	cache.Remove(calURL1, "/shared.ics")

	if cache.Get(calURL1, "/shared.ics") != nil {
		t.Error("Entry should be removed from cal1")
	}
	if cache.Get(calURL2, "/shared.ics") == nil {
		t.Error("Entry should still exist in cal2")
	}
}

func TestCacheEmptyData(t *testing.T) {
	dir := tempCacheDir(t)
	cache, err := NewCache(dir)
	if err != nil {
		t.Fatalf("NewCache() error = %v", err)
	}

	calURL := "https://example.com/cal/"
	entry := CacheEntry{
		Path:      "/event.ics",
		ETag:      "etag",
		Data:      "",
		FetchedAt: time.Now(),
	}

	cache.Put(calURL, entry)

	got := cache.Get(calURL, entry.Path)
	if got == nil {
		t.Fatal("Get() returned nil")
	}
	if got.Data != "" {
		t.Errorf("Data = %q, want empty", got.Data)
	}
}
