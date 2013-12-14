package main

import (
	"fmt"
	"github.com/codegangsta/martini"
	"github.com/codegangsta/martini-contrib/render"
	"github.com/codegangsta/martini-contrib/sessions"
	"github.com/joho/godotenv"
	"html/template"
	"labix.org/v2/mgo"
	"labix.org/v2/mgo/bson"
	"log"
	"net/http"
	"os"
	"os/signal"
	"regexp"
	"time"
)

//
// Models
//
const ENTRY_COLLECTION_NAME = "entries"
const USER_COLLECTION_NAME = "users"

type Entry struct {
	Id     bson.ObjectId `bson:"_id"`
	Date   string        `bson:"date"`
	Body   string        `bson:"body"`
	UserId bson.ObjectId `bson:"user_id"`
}

type User struct {
	Id   bson.ObjectId `bson:"_id"`
	Uid  string        `bson:"uid"`
	Name string        `bson:"name"`
}

//
// Handlers
//
const SESSION_USER_ID_KEY string = "user-id"

func authorize(w http.ResponseWriter, r *http.Request, db *mgo.Database, c martini.Context, session sessions.Session) {
	userId := session.Get(SESSION_USER_ID_KEY)
	if userId == nil {
		log.Println("Unauthorized access")
		http.Redirect(w, r, "/auth", http.StatusFound)
		return
	}

	var user User
	err := db.C(USER_COLLECTION_NAME).FindId(bson.ObjectIdHex(userId.(string))).One(&user)
	if err != nil {
		log.Println("User not found")
		http.Redirect(w, r, "/auth", http.StatusFound)
		return
	}

	c.Map(&user)
}

func validateDate(w http.ResponseWriter, r *http.Request, params martini.Params) {
	date := params["date"]
	matched, err := regexp.MatchString("[0-9]+-[0-9]+-[0-9]+", date)
	if err != nil {
		panic(err)
	}
	if !matched {
		log.Println("Invalid date: %s", date)
		http.Redirect(w, r, "/", http.StatusFound)
		return
	}
}

func rootHandler(w http.ResponseWriter, r *http.Request) {
	now := time.Now()
	today := fmt.Sprintf("%d-%d-%d", now.Year(), now.Month(), now.Day())
	http.Redirect(w, r, "/entries/"+today, http.StatusFound)
}

// TODO: Reduce arguments. Can't I redirect without w and req?
func showEntry(w http.ResponseWriter, r *http.Request, ren render.Render, db *mgo.Database, params martini.Params, user *User) {
	date := params["date"]
	var entry Entry
	err := db.C(ENTRY_COLLECTION_NAME).Find(bson.M{"user_id": user.Id, "date": date}).One(&entry)
	if err != nil {
		http.Redirect(w, r, "/entries/"+date+"/edit", http.StatusFound)
		return
	}
	data := make(map[string]interface{})
	data["Entry"] = &entry
	data["CurrentUser"] = user
	ren.HTML(200, "view", data)
}

func editEntry(r render.Render, db *mgo.Database, params martini.Params, user *User) {
	date := params["date"]
	var entry Entry
	err := db.C(ENTRY_COLLECTION_NAME).Find(bson.M{"user_id": user.Id, "date": date}).One(&entry)
	if err != nil {
		entry = Entry{Id: bson.NewObjectId(), Date: date, Body: "", UserId: user.Id}
	}
	data := make(map[string]interface{})
	data["Entry"] = &entry
	data["CurrentUser"] = user
	r.HTML(200, "edit", data)
}

func saveEntry(w http.ResponseWriter, r *http.Request, db *mgo.Database, params martini.Params, user *User) {
	date := params["date"]
	query := bson.M{"date": date, "user_id": user.Id}
	entry := bson.M{"date": date, "user_id": user.Id, "body": r.FormValue("body")}
	_, err := db.C(ENTRY_COLLECTION_NAME).Upsert(query, entry)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/entries/"+date, http.StatusFound)
}

func prepareRouter(m *martini.ClassicMartini) {
	m.Get("/", authorize, rootHandler)

	m.Get("/auth", authHandler)
	m.Get("/auth/logout", authLogoutHandler)
	m.Get("/auth/callback", getAccessToken, getUserInfo, findOrCreateUser)

	m.Get("/entries/:date", authorize, validateDate, showEntry)
	m.Post("/entries/:date", authorize, validateDate, saveEntry)
	m.Get("/entries/:date/edit", authorize, validateDate, editEntry)
}

//
// Execution
//
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
