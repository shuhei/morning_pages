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
	m.Get("/", authorize, rootHandler)

	m.Get("/auth", showLogin)
	m.Get("/auth/logout", logout)
	m.Get("/auth/callback", getAccessToken, getUserInfo, findOrCreateUser)

	m.Get("/entries/:date", authorize, validateDate, fetchDateEntries, showEntry)
	m.Post("/entries/:date", authorize, validateDate, saveEntry)
	m.Get("/entries/:date/edit", authorize, validateDate, fetchDateEntries, editEntry)
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
	sessionKey := os.Getenv("SESSION_KEY")
	store := sessions.NewCookieStore([]byte(sessionKey))
	m.Use(sessions.Sessions("default-session", store))

	//
	// Template
	//
	m.Use(render.Renderer(render.Options{
		Directory:  "templates",
		Extensions: []string{".html"},
		Funcs:      templateFuncs(),
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
	// TemplateData
	//
	m.Use(initTemplateData)

	//
	// Router
	//
	prepareRouter(m)

	m.Run()
}
