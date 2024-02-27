// Package conteo_traefik_emergency_cache is a plugin to cache responses to disk.
package conteo_traefik_emergency_cache

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"strings"

	// "regexp"

	"time"

	"github.com/igoooor/conteo-traefik-emergency-cache/provider/api"
)

// Config configures the middleware.
type Config struct {
	EmergencyMode   bool   `json:"emergencyMode" yaml:"emergencyMode" toml:"emergencyMode"`
	Path            string `json:"path" yaml:"path" toml:"path"`
	BypassHeader    string `json:"bypassHeader" yaml:"bypassHeader" toml:"bypassHeader"`
	CacheableHeader string `json:"cacheableHeader" yaml:"cacheableHeader" toml:"cacheableHeader"`
	Debug           bool   `json:"debug" yaml:"debug" toml:"debug"`
}

// CreateConfig returns a config instance.
func CreateConfig() *Config {
	return &Config{
		EmergencyMode:   false,
		BypassHeader:    "X-Emergency-Cache-Control",
		CacheableHeader: "X-Emergency-Cacheable",
		Debug:           false,
	}
}

type CacheSystem interface {
	Get(string) ([]byte, error)
	Set(string, []byte) error
}

type cache struct {
	name  string
	cache api.FileCache
	cfg   *Config
	next  http.Handler
}

// New returns a plugin instance.
func New(_ context.Context, next http.Handler, cfg *Config, name string) (http.Handler, error) {
	fc, err := api.NewFileCache(cfg.Path)
	if err != nil {
		return nil, err
	}

	m := &cache{
		name:  name,
		cache: *fc,
		cfg:   cfg,
		next:  next,
	}

	return m, nil
}

type cacheData struct {
	Status  int
	Headers map[string][]string
	Body    []byte
	Created uint64
}

// ServeHTTP serves an HTTP request.
func (m *cache) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if m.bypassingHeaders(r) {
		rw := &responseWriter{ResponseWriter: w}
		m.next.ServeHTTP(rw, r)

		return
	}

	key := m.cacheKey(r, true)

	if m.cfg.EmergencyMode {
		if m.cfg.Debug {
			log.Printf("[Emergency Cache] DEBUG get %s", key)
		}
		b, err := m.cache.Get(key)
		if m.processCachedResponse(r, w, b, err) {
			return
		}
		keyWithoutQueryParameters := m.cacheKey(r, false)
		if keyWithoutQueryParameters != key {
			if m.cfg.Debug {
				log.Printf("[Emergency Cache] DEBUG get (no query) %s", keyWithoutQueryParameters)
			}
			b, err := m.cache.Get(keyWithoutQueryParameters)
			if m.processCachedResponse(r, w, b, err) {
				return
			}
		}

		rw := &responseWriter{ResponseWriter: w}
		m.next.ServeHTTP(rw, r)
		return
	}

	if m.cfg.Debug {
		log.Printf("[Emergency Cache] DEBUG disabled")
	}

	rw := &responseWriter{ResponseWriter: w}
	m.next.ServeHTTP(rw, r)

	if !m.cacheable(r, w, rw.status) {
		if m.cfg.Debug {
			log.Printf("[Emergency Cache] DEBUG response not cacheable")
		}
		return
	}

	createdTs := uint64(time.Now().Unix())
	data := cacheData{
		Status:  rw.status,
		Headers: w.Header(),
		Body:    rw.body,
		Created: createdTs,
	}

	b, err := json.Marshal(data)
	if err != nil {
		log.Printf("Error serializing cache item: %v", err)
	}

	go m.cache.Set(key, b)

	if m.cfg.Debug {
		log.Printf("[Emergency Cache] DEBUG set %s", key)
	}
}

func (m *cache) processCachedResponse(r *http.Request, w http.ResponseWriter, b []byte, err error) bool {
	if err == nil && b != nil {
		var data cacheData

		err := json.Unmarshal(b, &data)
		if err == nil && data.Status == 200 {
			m.sendCacheFile(w, data, r)
			return true
		}
	} else if err != nil {
		log.Println(err)
	}

	return false
}

func (m *cache) cacheable(r *http.Request, w http.ResponseWriter, status int) bool {
	if m.cfg.Debug {
		log.Printf("[Emergency Cache] DEBUG cacheable?")
	}
	return status == 200 && (w.Header().Get(m.cfg.CacheableHeader) == "true" || strings.HasPrefix(r.URL.Path, "/build/"))
}

func (m *cache) sendCacheFile(w http.ResponseWriter, data cacheData, r *http.Request) {
	if m.cfg.Debug {
		log.Printf("[Emergency Cache] DEBUG hit")
	}

	for key, vals := range data.Headers {
		for _, val := range vals {
			w.Header().Add(key, val)
		}
	}

	w.WriteHeader(data.Status)
	_, _ = w.Write(data.Body)
}

func (m *cache) bypassingHeaders(r *http.Request) bool {
	return r.Header.Get(m.cfg.BypassHeader) == "no-cache"
}

func (m *cache) cacheKey(r *http.Request, addQuery bool) string {
	key := r.Host + r.URL.Path
	if r.URL.RawQuery != "" && addQuery {
		key += "?" + r.URL.RawQuery
	}

	if m.cfg.Debug {
		log.Printf("[Emergency Cache] DEBUG key: %s", key)
	}

	return key
}

type responseWriter struct {
	http.ResponseWriter
	status int
	body   []byte
}

func (rw *responseWriter) Header() http.Header {
	return rw.ResponseWriter.Header()
}

func (rw *responseWriter) Write(p []byte) (int, error) {
	rw.body = append(rw.body, p...)
	return rw.ResponseWriter.Write(p)
}

func (rw *responseWriter) WriteHeader(s int) {
	rw.status = s
	rw.ResponseWriter.WriteHeader(s)
}
