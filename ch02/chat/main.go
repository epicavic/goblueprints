package main

import (
	"flag"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"text/template"
	"time"

	"main/trace"

	"github.com/stretchr/gomniauth"
	"github.com/stretchr/gomniauth/providers/github"
	"github.com/stretchr/objx"
)

// templ represents a single template
type templateHandler struct {
	once     sync.Once
	filename string
	templ    *template.Template
}

// ServeHTTP handles the HTTP request.
func (t *templateHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	fmt.Println("templateHandler: ServeHTTP called")
	t.once.Do(func() {
		t.templ = template.Must(template.ParseFiles(filepath.Join("templates", t.filename)))
	})

	data := map[string]interface{}{
		"Host": r.Host,
	}
	if authCookie, err := r.Cookie("auth"); err == nil {
		data["UserData"] = objx.MustFromBase64(authCookie.Value)
	}

	t.templ.Execute(w, data)
}

const alnum = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

var host = flag.String("host", "localhost:8080", "The host of the application.")

// randSeq generates random sequence of alnum characters of fixed length
func randSeq(n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = alnum[rand.Intn(len(alnum))]
	}
	return string(b)
}

func main() {

	flag.Parse() // parse the flags
	githubCallbackURL := "http://" + *host + "/auth/callback/github"
	githubClientID := os.Getenv("GITHUB_CLIENT_ID")         // registered client ID
	githubClientSecret := os.Getenv("GITHUB_CLIENT_SECRET") // registered client secret
	gomniauthSecurityKey := os.Getenv("OAUTH_SECURITY_KEY") // random 64-bit secret key

	if gomniauthSecurityKey == "" {
		rand.Seed(time.Now().UnixNano())
		gomniauthSecurityKey = randSeq(64)
	}

	fmt.Println("githubCallbackURL:", githubCallbackURL)
	fmt.Println("githubClientID:", githubClientID)
	fmt.Println("githubClientSecret:", githubClientSecret)
	fmt.Println("gomniauthSecurityKey:", gomniauthSecurityKey)

	// setup gomniauth
	gomniauth.SetSecurityKey(gomniauthSecurityKey)
	gomniauth.WithProviders(
		github.New(githubClientID, githubClientSecret, githubCallbackURL),
	)

	r := newRoom()
	r.tracer = trace.New(os.Stdout)

	http.Handle("/chat", MustLogin(&templateHandler{filename: "chat.html"}))
	http.Handle("/login", &templateHandler{filename: "login.html"})
	http.HandleFunc("/auth/", authHandler)
	http.Handle("/room", r)

	// get the room going
	go r.run()

	// start the web server
	log.Println("Starting web server on", *host)
	if err := http.ListenAndServe(*host, nil); err != nil {
		log.Fatal("ListenAndServe:", err)
	}

}
