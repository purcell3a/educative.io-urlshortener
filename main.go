package main

import (
	"fmt"
	"net/http"
	"https://github.com/purcell3a/educative.io-urlshortener/store"
	"https://github.com/purcell3a/educative.io-urlshortener/key"

)

const addForm = `
<html><body>
<form method="POST" action="/add">
URL: <input type="text" name="url">
<input type="submit" value="Add">
</form>
</html></body>
`

var store = NewURLStore("store.gob")

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
	url := r.FormValue("url") //Read in the long URL, which means reading the URL from an HTML-form contained in an HTTP request with r.FormValue("url").
	if url == "" {
		fmt.Fprint(w, addForm)
		return
	}
	key := store.Put(url) //Put it in the store using our Put method on the store.
	//send the corresponding short URL to the user.
	fmt.Fprintf(w, "%s", key)
}

func main() {
	http.HandleFunc("/", Redirect)
	http.HandleFunc("/add", Add)
	http.ListenAndServe(":3000", nil)
}
