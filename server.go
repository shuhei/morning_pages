package main

import (
	"github.com/codegangsta/martini"
	"github.com/codegangsta/martini-contrib/render"
	"github.com/codegangsta/martini-contrib/sessions"
	"github.com/codegangsta/martini-contrib/web"
	"github.com/joho/godotenv"
	"labix.org/v2/mgo"
	"log"
	"os"
	"os/signal"
)

func prepareRouter(m *martini.ClassicMartini) {
	m.Get("/", Authorize, ShowRoot)

	m.Get("/auth", ShowLogin)
	m.Get("/auth/logout", Logout)
	m.Get("/auth/callback", GetAccessToken, GetUserInfo, FindOrCreateUser)

	m.Get("/entries", Authorize, GetEntries)
	m.Get("/entries/:date", Authorize, ValidateDate, GetEntry)
	m.Post("/entries/:date", Authorize, ValidateDate, CreateEntry)
	m.Put("/entries/:date", Authorize, ValidateDate, UpdateEntry)
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
	m.MapTo(&userStore{db}, (*UserStore)(nil))
	m.MapTo(&entryStore{db}, (*EntryStore)(nil))

	//
	// Session
	//
	sessionKey := os.Getenv("SESSION_KEY")
	store := sessions.NewCookieStore([]byte(sessionKey))
	m.Use(sessions.Sessions("default-session", store))

	//
	// Template
	//
	m.Use(render.Renderer(render.Options{
		Directory:  "templates",
		Extensions: []string{".html"},
		Layout:     "layout",
	}))

	//
	// Facebook Auth
	//
	appId := os.Getenv("FB_APP_ID")
	appSecret := os.Getenv("FB_APP_SECRET")
	redirectUrl := os.Getenv("FB_REDIRECT_URL")
	fb := NewFacebookAuth(appId, appSecret, redirectUrl)
	m.MapTo(fb, (*FacebookAuth)(nil))

	//
	// web.go context
	//
	m.Use(web.ContextWithCookieSecret(sessionKey))

	//
	// Router
	//
	prepareRouter(m)

	m.Run()
}
