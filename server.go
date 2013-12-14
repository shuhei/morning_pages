package main

import (
	"github.com/codegangsta/martini"
	"github.com/codegangsta/martini-contrib/render"
	"github.com/codegangsta/martini-contrib/sessions"
	"github.com/joho/godotenv"
	"html/template"
	"labix.org/v2/mgo"
	"log"
	"os"
	"os/signal"
)

func prepareRouter(m *martini.ClassicMartini) {
	m.Get("/", authorize, rootHandler)

	m.Get("/auth", showLogin)
	m.Get("/auth/logout", logout)
	m.Get("/auth/callback", getAccessToken, getUserInfo, findOrCreateUser)

	m.Get("/entries/:date", authorize, validateDate, showEntry)
	m.Post("/entries/:date", authorize, validateDate, saveEntry)
	m.Get("/entries/:date/edit", authorize, validateDate, editEntry)
}

// Execute cleanup func when the server is killed.
func cleanupBeforeExit(cleanup func()) {
	terminateChan := make(chan os.Signal, 1)
	signal.Notify(terminateChan, os.Interrupt)
	signal.Notify(terminateChan, os.Kill)
	go func() {
		<-terminateChan
		cleanup()
	}()
}

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Println("Failed to load .env. Maybe on production?")
	}

	m := martini.Classic()

	//
	// Database
	//
	session, err := mgo.Dial(os.Getenv("MONGOHQ_URL"))
	if err != nil {
		log.Fatal(err)
	}
	cleanupBeforeExit(func() {
		log.Println("Cleaning up...")
		session.Close()
		os.Exit(1)
	})

	db := session.DB("") // Use database specified in the URL.
	m.Map(db)

	//
	// Session
	//
	store := sessions.NewCookieStore([]byte(os.Getenv("SESSION_KEY")))
	m.Use(sessions.Sessions("default-session", store))

	//
	// Template
	//
	funcMap := template.FuncMap{
		"unsafe":    unsafe,
		"linebreak": linebreak,
		"charCount": charCount,
	}
	m.Use(render.Renderer(render.Options{
		Directory:  "templates",
		Extensions: []string{".html"},
		Funcs:      []template.FuncMap{funcMap},
		Layout:     "layout",
	}))

	//
	// Facebook Auth
	//
	appId := os.Getenv("FB_APP_ID")
	appSecret := os.Getenv("FB_APP_SECRET")
	fb := &FacebookAuth{AppId: appId, AppSecret: appSecret}
	m.Map(fb)

	//
	// Router
	//
	prepareRouter(m)

	m.Run()
}
