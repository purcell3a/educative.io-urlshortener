package main

import (
	"fmt"
	"net/http"
	"sync"
)

const addForm = `
<html><body>
<form method="POST" action="/add">
URL: <input type="text" name="url">
<input type="submit" value="Add">
</form>
</html></body>
`

var store = NewURLStore()

type URLStore struct {
	urls map[string]string // map from short to long URLs
	mu   sync.RWMutex  //An RWMutex has two locks: one for readers and one for writers
}

//In return, we make a URLStore literal with our map initialized. The lock doesn’t need to be specifically 
//initialized; this is the standard way in Go of making struct objects. & is the address-of operator,
// to return a pointer because NewURLStore returns the pointer *URLStore. We call this function to make a 
//URLStore variable store:
func NewURLStore() *URLStore {
	return &URLStore{ urls: make(map[string]string) }
}


func (s *URLStore) Get(key string) string{ // this fun locks and unlocks to keep processes from
	defer s.mu.RLock()						  // being interrupted by requests
	return s.urls[key]
}

func (s *URLStore) Set(key, url string) bool{ //The Set function needs both a key and a URL and has to
	s.mu.Lock()		
	defer s.mu.Unlock()						//use the write lock Lock() to exclude any other updates at the 
	_, present := s.urls[key]				//same time. It returns a boolean true or false value to indicate 
	if present {							//whether the Set was successful or not:
		return false
	}
	s.urls[key] = url 
	return true
}


func main() {
	http.HandleFunc("/", Redirect)
	http.HandleFunc("/add", Add)
	http.ListenAndServe(":3000", nil)
}

func Redirect(w http.ResponseWriter, r *http.Request) {
	key := r.URL.Path[1:]
	url := store.Get(key)
	if url == "" {
		http.NotFound(w, r)
		return
	}
	http.Redirect(w, r, url, http.StatusFound)
}

func Add(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	url := r.FormValue("url")
	if url == "" {
		fmt.Fprint(w, addForm)
		return
	}
	key := store.Put(url)

	fmt.Fprintf(w, "%s", key)
}
