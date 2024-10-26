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
}

func Run(port, origin string) error {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// Validation
		if r.Method != http.MethodGet {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("Method " + r.Method + " not allowed\n"))
			log.Println("unsupported method.")
			return
		}

		t := time.Now()

		// Search in cache
		rKey := prepareCacheKey(r)
		wg.RLock()
		if cachedResp, ok := cache[rKey]; ok {
			sendCachedResponse(w, cachedResp)
			wg.RUnlock()
			log.Printf("response took from cache. Elapsed in %dms\n", time.Since(t).Milliseconds())
			return
		}
		wg.RUnlock()

		// Sending request to origin
		log.Printf("sending request to: %s", origin+r.URL.Path)

		resp, err := http.Get(origin + r.URL.Path)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer resp.Body.Close()

		// Send response
		sendResp(w, resp)

		// Cache response
		wg.Lock()
		cache[rKey] = newCacheResp(resp)
		wg.Unlock()
		log.Printf("response cached. Elapsed in %dms\n", time.Since(t).Milliseconds())
	})

	if err := http.ListenAndServe(":"+port, nil); err != nil && err != http.ErrServerClosed {
		log.Fatal(err)
	}

	return nil
}

// sendResp sends a response to the client, copying the headers and body
func sendCachedResponse(w http.ResponseWriter, ci *cacheItem) {
	w.WriteHeader(ci.StatusCode)

	for key, values := range ci.Header {
		for _, value := range values {
			w.Header().Add(key, value)
		}
	}

	w.Header().Set("X-Cache", "HIT")

	if _, err := w.Write(ci.Body); err != nil {
		log.Fatalf("error while writing response body: %v", err)
	}

	return
}

// prepareCacheKey creates a unique cache key for the request
func prepareCacheKey(r *http.Request) string {
	var keyBuilder strings.Builder

	keyBuilder.WriteString(r.URL.Path)

	for key, values := range r.Header {
		_, err := keyBuilder.WriteString(key)
		if err != nil {
			log.Fatal(err)
		}

		_, err = keyBuilder.WriteString(strings.Join(values, ""))
		if err != nil {
			log.Fatal(err)
		}
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		log.Fatalf("error reading request body: %v", err)
	}
	keyBuilder.Write(body)

	r.Body = io.NopCloser(bytes.NewReader(body))

	hash := sha256.New()
	hash.Write([]byte(keyBuilder.String()))
	return hex.EncodeToString(hash.Sum(nil))
}

func newCacheResp(r *http.Response) *cacheItem {
	ci := &cacheItem{
		StatusCode: r.StatusCode,
		Header:     make(http.Header),
	}

	for key, valuse := range r.Header {
		for _, value := range valuse {
			ci.Header.Add(key, value)
		}
	}
	ci.Header.Add("X-Cache", "HIT")

	body, err := io.ReadAll(r.Body)
	if err != nil {
		log.Fatal(err)
	}

	ci.Body = body

	return ci
}

func sendResp(w http.ResponseWriter, r *http.Response) {
	w.WriteHeader(r.StatusCode)

	for key, valuse := range r.Header {
		for _, value := range valuse {
			w.Header().Add(key, value)
		}
	}
	w.Header().Add("X-Cache", "MISS")

	body, err := io.ReadAll(r.Body)
	if err != nil {
		log.Fatal(err)
	}

	r.Body = io.NopCloser(bytes.NewReader(body))

	w.Write(body)

	return
}
