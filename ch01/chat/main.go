package main

import (
	"flag"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"text/template"

	"main/trace"
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
	t.once.Do(func() {
		t.templ = template.Must(template.ParseFiles(filepath.Join("templates", t.filename)))
	})
	t.templ.Execute(w, r)
}

func main() {
	var addr = flag.String("addr", "localhost:8080", "The addr of the application.")
	flag.Parse() // parse the flags

	r := newRoom()
	r.tracer = trace.New(os.Stdout)

	http.Handle("/", &templateHandler{filename: "chat.html"})
	http.Handle("/room", r)

	// get the room going
	go r.run()

	// start the web server
	log.Println("Starting web server on", *addr)
	log.Fatal(http.ListenAndServe(*addr, nil))
}
