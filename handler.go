package main

import (
	"fmt"
	"github.com/codegangsta/martini"
	"github.com/codegangsta/martini-contrib/render"
	"github.com/codegangsta/martini-contrib/sessions"
	"labix.org/v2/mgo"
	"log"
	"net/http"
	"regexp"
	"time"
)

const SESSION_USER_ID_KEY string = "user-id"

//
// Filters
//

func authorize(w http.ResponseWriter, r *http.Request, db *mgo.Database, c martini.Context, session sessions.Session, l *log.Logger) {
	userId := session.Get(SESSION_USER_ID_KEY)
	if userId == nil {
		l.Println("Unauthorized access")
		http.Redirect(w, r, "/auth", http.StatusFound)
		return
	}

	user, err := findUserById(db, userId.(string))
	if err != nil {
		l.Println("User not found")
		http.Redirect(w, r, "/auth", http.StatusFound)
		return
	}

	c.Map(user)
}

func validateDate(w http.ResponseWriter, r *http.Request, params martini.Params) {
	date := params["date"]
	matched, err := regexp.MatchString("[0-9]+-[0-9]+-[0-9]+", date)
	if err != nil {
		panic(err)
	}
	if !matched {
		log.Println("Invalid date:", date)
		http.Redirect(w, r, "/", http.StatusFound)
		return
	}
}

//
// Handlers
//

func rootHandler(w http.ResponseWriter, r *http.Request) {
	now := time.Now()
	today := fmt.Sprintf("%d-%d-%d", now.Year(), now.Month(), now.Day())
	http.Redirect(w, r, "/entries/"+today, http.StatusFound)
}

// TODO: Reduce arguments. Can't I redirect without w and req?
func showEntry(w http.ResponseWriter, r *http.Request, ren render.Render, db *mgo.Database, params martini.Params, user *User) {
	date := params["date"]
	entry, err := findEntry(db, user, date)
	if err != nil {
		http.Redirect(w, r, "/entries/"+date+"/edit", http.StatusFound)
		return
	}
	data := make(map[string]interface{})
	data["Entry"] = entry
	data["CurrentUser"] = user
	ren.HTML(200, "view", data)
}

func editEntry(r render.Render, db *mgo.Database, params martini.Params, user *User) {
	date := params["date"]
	entry, err := findEntry(db, user, date)
	if err != nil {
		entry = newEntry(user, date)
	}
	data := make(map[string]interface{})
	data["Entry"] = entry
	data["CurrentUser"] = user
	r.HTML(200, "edit", data)
}

func saveEntry(w http.ResponseWriter, r *http.Request, db *mgo.Database, params martini.Params, user *User) {
	date := params["date"]
	body := r.FormValue("body")
	err := upsertEntry(db, user, date, body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/entries/"+date, http.StatusFound)
}
