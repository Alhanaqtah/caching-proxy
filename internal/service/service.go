package service

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"io"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"
)

var (
	cache = make(map[string]*cacheItem)
	wg    sync.RWMutex
)

type cacheItem struct {
	StatusCode int
	Header     http.Header
	Body       []byte
	Timestamp  time.Time
}

// Run starts the caching proxy server on the specified port, forwarding requests to the origin server and caching responses
func Run(port, origin string, ttl time.Duration) error {
	go cleanCacheWorker(ttl)

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// Validate request method
		if r.Method != http.MethodGet {
			http.Error(w, "Method "+r.Method+" not allowed\n", http.StatusBadRequest)
			log.Println("Unsupported method.")
			return
		}

		t := time.Now()

		// Search for response in cache
		rKey, err := prepareCacheKey(r)
		if err != nil {
			http.Error(w, "Error preparing cache key: "+err.Error(), http.StatusInternalServerError)
			return
		}

		wg.RLock()
		if cachedResp, ok := cache[rKey]; ok {
			sendCachedResponse(w, cachedResp)
			wg.RUnlock()

			log.Printf("Response served from cache. Elapsed in %dms\n", time.Since(t).Milliseconds())
			return
		}
		wg.RUnlock()

		// Forward request to origin server
		log.Printf("Sending request to: %s", origin+r.URL.Path)
		resp, err := http.Get(origin + r.URL.Path)
		if err != nil {
			http.Error(w, "Error contacting origin server: "+err.Error(), http.StatusInternalServerError)
			return
		}
		defer resp.Body.Close()

		// Send response to client and cache it
		sendResp(w, resp)

		ci, err := newCacheResp(resp)
		if err != nil {
			http.Error(w, "Error caching response: "+err.Error(), http.StatusInternalServerError)
			return
		}

		wg.Lock()
		cache[rKey] = ci
		wg.Unlock()
		log.Printf("Response cached. Elapsed in %dms\n", time.Since(t).Milliseconds())
	})

	if err := http.ListenAndServe(":"+port, nil); err != nil && err != http.ErrServerClosed {
		log.Printf("Error starting server: %v", err)
		return err
	}

	return nil
}

// sendCachedResponse writes the cached response to the client
func sendCachedResponse(w http.ResponseWriter, ci *cacheItem) {
	w.WriteHeader(ci.StatusCode)
	for key, values := range ci.Header {
		for _, value := range values {
			w.Header().Add(key, value)
		}
	}
	w.Header().Set("X-Cache", "HIT")

	if _, err := w.Write(ci.Body); err != nil {
		log.Printf("Error while writing response body: %v", err)
	}
}

// prepareCacheKey creates a unique cache key based on the request URL and headers
func prepareCacheKey(r *http.Request) (string, error) {
	var keyBuilder strings.Builder

	keyBuilder.WriteString(r.URL.Path)
	for key, values := range r.Header {
		keyBuilder.WriteString(key)
		keyBuilder.WriteString(strings.Join(values, ""))
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		log.Printf("Error reading request body: %v", err)
		return "", err
	}
	defer r.Body.Close()

	keyBuilder.Write(body)
	r.Body = io.NopCloser(bytes.NewReader(body))

	hash := sha256.New()
	hash.Write([]byte(keyBuilder.String()))
	return hex.EncodeToString(hash.Sum(nil)), nil
}

// newCacheResp creates a cache item from the response and returns it
func newCacheResp(r *http.Response) (*cacheItem, error) {
	ci := &cacheItem{
		StatusCode: r.StatusCode,
		Header:     make(http.Header),
		Timestamp:  time.Now(),
	}

	for key, values := range r.Header {
		for _, value := range values {
			ci.Header.Add(key, value)
		}
	}
	ci.Header.Add("X-Cache", "HIT")

	body, err := io.ReadAll(r.Body)
	if err != nil {
		log.Printf("Error reading body: %v", err)
		return nil, err
	}

	ci.Body = body
	return ci, nil
}

// sendResp sends the response from the origin server to the client
func sendResp(w http.ResponseWriter, r *http.Response) {
	w.WriteHeader(r.StatusCode)
	for key, values := range r.Header {
		for _, value := range values {
			w.Header().Add(key, value)
		}
	}
	w.Header().Set("X-Cache", "MISS")

	body, err := io.ReadAll(r.Body)
	if err != nil {
		log.Printf("Error reading response body: %v", err)
		return
	}

	r.Body = io.NopCloser(bytes.NewReader(body))
	w.Write(body)
}

// cleanCacheWorker periodically removes expired items from the cache based on the specified TTL
func cleanCacheWorker(ttl time.Duration) {
	ticker := time.NewTicker(ttl)
	defer ticker.Stop()

	for range ticker.C {
		wg.Lock()
		for key, item := range cache {
			if time.Since(item.Timestamp) > ttl {
				delete(cache, key)
			}
		}
		wg.Unlock()
	}
}
