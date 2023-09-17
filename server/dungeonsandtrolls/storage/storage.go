// Persistent storage backed by a file.

package storage

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"
)

type Storage struct {
	path string
	lock sync.RWMutex
	data map[string]string
}

func NewStorage(path string) (*Storage, error) {
	_, err := os.Stat(path)
	var j []byte
	if os.IsNotExist(err) {
		j = []byte("{}")
	} else {
		j, err = os.ReadFile(path)
		if err != nil {
			return nil, err
		}
	}
	s := &Storage{
		path: path,
	}
	err = json.Unmarshal(j, &s.data)
	return s, err
}

func (s *Storage) write() error {
	j, err := json.Marshal(s.data)
	if err != nil {
		return err
	}
	tempFile := s.path + ".tmp"
	err = os.WriteFile(s.path, j, 0644)
	if err != nil {
		return err
	}
	return os.Rename(tempFile, s.path)
}

// Write a value (has to be serializable to JSON) to the storage identified by the key.
func (s *Storage) Write(key string, value any) error {
	j, err := json.Marshal(value)
	if err != nil {
		return err
	}
	s.lock.Lock()
	defer s.lock.Unlock()
	s.data[key] = string(j)
	return s.write()
}

func (s *Storage) Read(key string) (any, error) {
	s.lock.RLock()
	defer s.lock.RUnlock()
	jd, ok := s.data[key]
	if !ok {
		return nil, fmt.Errorf("key %s not found in %s", key, s.path)
	}
	var v interface{}
	err := json.Unmarshal([]byte(jd), &v)
	return v, err
}

func (s *Storage) ReadTo(key string, v any) error {
	s.lock.RLock()
	defer s.lock.RUnlock()
	jd, ok := s.data[key]
	if !ok {
		return fmt.Errorf("key %s not found in %s", key, s.path)
	}
	return json.Unmarshal([]byte(jd), &v)
}
