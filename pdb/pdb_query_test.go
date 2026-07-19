package pdb

import (
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"sync/atomic"
	"testing"
	"time"

	"pdb_proxy/conf"

	"github.com/gofiber/fiber/v2"
)

func TestCacheIsFresh(t *testing.T) {
	originalTTL := conf.PdbCacheTTL
	t.Cleanup(func() { conf.PdbCacheTTL = originalTTL })

	now := time.Date(2026, time.January, 1, 12, 0, 0, 0, time.UTC)
	fileInfo := testFileInfo{modTime: now.Add(-30 * time.Minute)}

	conf.PdbCacheTTL = 0
	if !cacheIsFresh(fileInfo, now) {
		t.Fatal("zero TTL should keep the cache fresh forever")
	}

	conf.PdbCacheTTL = time.Hour
	if !cacheIsFresh(fileInfo, now) {
		t.Fatal("cache within TTL should be fresh")
	}

	fileInfo.modTime = now.Add(-2 * time.Hour)
	if cacheIsFresh(fileInfo, now) {
		t.Fatal("cache past TTL should be stale")
	}
}

func TestPdbQueryCacheTTL(t *testing.T) {
	originalDir, originalServer, originalTTL := conf.PdbDir, conf.PdbServer, conf.PdbCacheTTL
	t.Cleanup(func() {
		conf.PdbDir = originalDir
		conf.PdbServer = originalServer
		conf.PdbCacheTTL = originalTTL
	})

	var requests atomic.Int32
	content := atomic.Value{}
	content.Store("version-1")
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		requests.Add(1)
		_, _ = io.WriteString(w, content.Load().(string))
	}))
	defer upstream.Close()

	conf.PdbDir = t.TempDir()
	conf.PdbServer = upstream.URL
	conf.PdbCacheTTL = time.Hour

	app := fiber.New()
	app.Get("/download/symbols/:pdbname/:pdbhash/:pdbname", PdbQuery)
	requestURL := "/download/symbols/test.pdb/HASH/test.pdb"

	assertBody := func(want string) {
		t.Helper()
		response, err := app.Test(httptest.NewRequest(http.MethodGet, requestURL, nil))
		if err != nil {
			t.Fatal(err)
		}
		defer response.Body.Close()
		body, err := io.ReadAll(response.Body)
		if err != nil {
			t.Fatal(err)
		}
		if response.StatusCode != http.StatusOK || string(body) != want {
			t.Fatalf("status=%d body=%q, want status=200 body=%q", response.StatusCode, body, want)
		}
	}

	assertBody("version-1")
	assertBody("version-1")
	if got := requests.Load(); got != 1 {
		t.Fatalf("fresh cache made %d upstream requests, want 1", got)
	}

	cachePath := filepath.Join(conf.PdbDir, "test.pdb", "HASH", "test.pdb")
	oldTime := time.Now().Add(-2 * time.Hour)
	if err := os.Chtimes(cachePath, oldTime, oldTime); err != nil {
		t.Fatal(err)
	}
	content.Store("version-2")
	assertBody("version-2")
	if got := requests.Load(); got != 2 {
		t.Fatalf("expired cache made %d upstream requests, want 2", got)
	}

	conf.PdbCacheTTL = 0
	oldTime = time.Now().Add(-24 * time.Hour)
	if err := os.Chtimes(cachePath, oldTime, oldTime); err != nil {
		t.Fatal(err)
	}
	content.Store("version-3")
	assertBody("version-2")
	if got := requests.Load(); got != 2 {
		t.Fatalf("permanent cache made %d upstream requests, want 2", got)
	}
}

func TestPdbQueryServesStaleCacheWhenRefreshFails(t *testing.T) {
	originalDir, originalServer, originalTTL := conf.PdbDir, conf.PdbServer, conf.PdbCacheTTL
	t.Cleanup(func() {
		conf.PdbDir = originalDir
		conf.PdbServer = originalServer
		conf.PdbCacheTTL = originalTTL
	})

	conf.PdbDir = t.TempDir()
	conf.PdbCacheTTL = time.Hour
	cachePath := filepath.Join(conf.PdbDir, "test.pdb", "HASH", "test.pdb")
	if err := os.MkdirAll(filepath.Dir(cachePath), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(cachePath, []byte("stale-but-valid"), 0644); err != nil {
		t.Fatal(err)
	}
	oldTime := time.Now().Add(-2 * time.Hour)
	if err := os.Chtimes(cachePath, oldTime, oldTime); err != nil {
		t.Fatal(err)
	}

	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		http.Error(w, "unavailable", http.StatusServiceUnavailable)
	}))
	defer upstream.Close()
	conf.PdbServer = upstream.URL

	app := fiber.New()
	app.Get("/download/symbols/:pdbname/:pdbhash/:pdbname", PdbQuery)
	response, err := app.Test(httptest.NewRequest(http.MethodGet, "/download/symbols/test.pdb/HASH/test.pdb", nil))
	if err != nil {
		t.Fatal(err)
	}
	defer response.Body.Close()
	body, err := io.ReadAll(response.Body)
	if err != nil {
		t.Fatal(err)
	}
	if response.StatusCode != http.StatusOK || string(body) != "stale-but-valid" {
		t.Fatalf("status=%d body=%q", response.StatusCode, body)
	}
}

type testFileInfo struct {
	modTime time.Time
}

func (f testFileInfo) Name() string       { return "test.pdb" }
func (f testFileInfo) Size() int64        { return 0 }
func (f testFileInfo) Mode() os.FileMode  { return 0 }
func (f testFileInfo) ModTime() time.Time { return f.modTime }
func (f testFileInfo) IsDir() bool        { return false }
func (f testFileInfo) Sys() any           { return nil }
