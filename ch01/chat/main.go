package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"path/filepath"
	"sync"
	"text/template"
)

// templ represents a single template
type templateHandler struct {
	once     sync.Once // ensure that template compiled only once.
	filename string
	templ    *template.Template
}

// ServeHTTP handles the HTTP request.
// ServeHTTP would be called multiple times by multiple http clients
// but template would be compiled only once and executed multiple times.
// An alternative would be some initialization code from main goroutine
// which would compile template once (NewTemplateHandler function or alike).
func (t *templateHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	fmt.Println("templateHandler: ServeHTTP called")
	t.once.Do(func() {
		t.templ = template.Must(template.ParseFiles(filepath.Join("templates", t.filename)))
	})
	t.templ.Execute(w, r)
}

func main() {
	var addr = flag.String("addr", "localhost:8080", "The addr of the application.")
	flag.Parse() // parse the flags

	// create a room instance
	r := newRoom()

	http.Handle("/", &templateHandler{filename: "chat.html"})
	http.Handle("/room", r) // called from javascript code when creating socket

	// get the room going
	go r.run()

	// start the web server
	log.Println("Starting web server on", *addr)
	log.Fatal(http.ListenAndServe(*addr, nil))
}

/*
$ go run .
2021/05/07 15:15:09 Starting web server on localhost:8080

## open first tab to localhost:8080
templateHandler: ServeHTTP called
room: ServeHTTP called
New client joined:  &{0xc0001389a0 0xc000068720 0xc00007ed50}
templateHandler: ServeHTTP called

## open second tab to localhost:8080
templateHandler: ServeHTTP called
room: ServeHTTP called
New client joined:  &{0xc0000ac160 0xc00008e120 0xc00007ed50}
templateHandler: ServeHTTP called

## send message from first tab
Message received:  hello from client1
 -- sent to client:  &{0xc0001389a0 0xc000068720 0xc00007ed50}
 -- sent to client:  &{0xc0000ac160 0xc00008e120 0xc00007ed50}

## send message from second tab
Message received:  hello from client2
 -- sent to client:  &{0xc0001389a0 0xc000068720 0xc00007ed50}
 -- sent to client:  &{0xc0000ac160 0xc00008e120 0xc00007ed50}

## open third tab to localhost:8080/room
room: ServeHTTP called
room: ServeHTTP: websocket: the client is not using the websocket protocol: 'upgrade' token not found in 'Connection' header
templateHandler: ServeHTTP called
*/
