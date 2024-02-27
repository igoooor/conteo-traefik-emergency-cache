// Package api is a api cache
package api

import (
	"encoding/base64"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
)

// Cache DB implementation
type FileCache struct {
	path string
}

// NewFileCache creates a new FileCache instance.
func NewFileCache(path string) (*FileCache, error) {
	fc := &FileCache{
		path: strings.TrimSuffix(path, "/") + "/",
	}

	return fc, nil
}

func encodeKey(key string) string {
	return base64.URLEncoding.EncodeToString([]byte(key))
}

// Get returns the value for the given key.
func (c *FileCache) Get(key string) ([]byte, error) {
	req, err := http.NewRequest(http.MethodGet, c.path+encodeKey(key), nil)
	if err != nil {
		return nil, err
	}

	client := &http.Client{}

	response, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	if response.StatusCode != http.StatusOK {
		return nil, nil
	}

	responseData, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}

	return responseData, nil
}

// Set sets the value for the given key.
func (c *FileCache) Set(key string, val []byte) error {
	req, err := http.NewRequest(http.MethodPut, c.path+encodeKey(key), strings.NewReader(string(val)))
	if err != nil {
		log.Printf("Error setting cache item for key %s", key)
		log.Println(err)
		return err
	}

	client := &http.Client{}

	_, err = client.Do(req)
	if err != nil {
		log.Printf("Error setting cache item for key %s", key)
		log.Println(err)
		return err
	}

	return nil
}
