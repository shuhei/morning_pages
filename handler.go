package main

import (
	"fmt"
	"github.com/codegangsta/martini"
	"github.com/codegangsta/martini-contrib/render"
	"github.com/codegangsta/martini-contrib/sessions"
	"labix.org/v2/mgo"
	"log"
	"net/http"
	"time"
)

const SESSION_USER_ID_KEY string = "user-id"

func dateString(t time.Time) string {
	return fmt.Sprintf("%04d-%02d-%02d", t.Year(), t.Month(), t.Day())
}

func isValidDate(date string) bool {
	_, err := time.ParseInLocation("2006-01-02", date, time.Local)
	return err == nil
}

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
	if !isValidDate(date) {
		http.Error(w, "Invalid date. e.g. 2014-01-02", http.StatusBadRequest)
		return
	}
}

//
// Handlers
//

func rootHandler(w http.ResponseWriter, r *http.Request) {
	today := dateString(time.Now())
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
	entries, err := findEntries(db, user)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	data := make(map[string]interface{})
	data["Entry"] = entry
	data["Entries"] = entries
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
