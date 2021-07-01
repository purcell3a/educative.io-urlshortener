package main

import (
	"encoding/gob"
	"io"
	"log"
	"os"
	"sync"
)

type record struct {
	Key, URL string
}

type URLStore struct {
	urls map[string]string // map from short to long URLs
	mu   sync.RWMutex      //An RWMutex has two locks: one for readers and one for writers
	file *os.File
}

// The NewURLStore function now takes a filename argument, opens the
// file, and stores the *os.File value in the file field of our
// URLStore variable store, here, locally called s.
func NewURLStore(filename string) *URLStore {
	s := &URLStore{urls: make(map[string]string)}
	f, err := os.OpenFile(filename, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		log.Fatal("Error opening URLStore:", err)
	}
	s.file = f
	if err := s.load(); err != nil {
		log.Println("Error loading data in URLStore:", err)
	}
	return s
}

func (s *URLStore) save(key, url string) error {
	e := gob.NewEncoder(s.file)
	return e.Encode(record{key, url})
}

// The new load method will seek to the beginning of the file,
// read and decode each record, and store the data in the map
// using the Set method. Again, notice the all-pervasive
// error-handling. The decoding of the file is an infinite
// loop which continues as long as there is no error:
func (s *URLStore) load() error {
	if _, err := s.file.Seek(0, 0); err != nil {
		return err
	}
	d := gob.NewDecoder(s.file)
	var err error
	for err == nil {
		var r record
		if err = d.Decode(&r); err == nil {
			s.Set(r.Key, r.URL)
		}
	}
	if err == io.EOF {
		return nil
	}
	return err
}

func (s *URLStore) Get(key string) string { // this fun locks and unlocks to keep processes from
	defer s.mu.RLock() // being interrupted by requests
	return s.urls[key]
}

func (s *URLStore) Put(url string) string {
	//The for loop retries the Set until it is successful (meaning that we have generated
	//a not yet existing short URL).
	for {
		key := genKey(s.Count())
		if s.Set(key, url) {
			if err := s.save(key, url); err != nil {
				log.Println("Error saving to URLStore:", err)
			}
			return key
		}
	}
	return ""
}

func (s *URLStore) Set(key, url string) bool { //The Set function needs both a key and a URL and has to
	s.mu.Lock()
	defer s.mu.Unlock()       //use the write lock Lock() to exclude any other updates at the
	_, present := s.urls[key] //same time. It returns a boolean true or false value to indicate
	if present {              //whether the Set was successful or not:
		return false
	}
	s.urls[key] = url
	return true
}

func (s *URLStore) Count() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.urls)
}
