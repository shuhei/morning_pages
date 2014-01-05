package main

import (
	"encoding/json"
	"github.com/codegangsta/martini"
	"github.com/codegangsta/martini-contrib/render"
	"github.com/codegangsta/martini-contrib/sessions"
	"github.com/codegangsta/martini-contrib/web"
	"io/ioutil"
	"labix.org/v2/mgo"
	"log"
	"net/http"
	"time"
)

const SessionUserIdKey string = "user-id"

//
// Filters
//

func authorize(ctx *web.Context, db *mgo.Database, c martini.Context, session sessions.Session, l *log.Logger) {
	userId := session.Get(SessionUserIdKey)
	if userId == nil || userId == "" {
		l.Println("Unauthorized access")
		ctx.Redirect(http.StatusFound, "/auth")
		return
	}

	user, err := findUserById(db, userId.(string))
	if err != nil {
		l.Println("User not found")
		session.Delete(SessionUserIdKey)
		ctx.Redirect(http.StatusFound, "/auth")
		return
	}

	c.Map(user)
}

func validateDate(ctx *web.Context, params martini.Params) {
	date := params["date"]
	if !isValidDate(date) {
		ctx.Abort(http.StatusBadRequest, "Invalid date. e.g. 2014-01-02")
		return
	}
}

//
// Handlers
//

func rootHandler(ren render.Render, user *User) {
	data := make(map[string]interface{})
	data["CurrentUser"] = user
	ren.HTML(200, "view", data)
}

//
// JSON APIs
//

func showEntry(ren render.Render, db *mgo.Database, params martini.Params, user *User) {
	date := params["date"]
	entry, err := findEntry(db, user, date)
	if err != nil {
		entry = newEntry(user, date)
	}
	ren.JSON(200, entry)
}

func saveEntry(ctx *web.Context, ren render.Render, db *mgo.Database, params martini.Params, user *User, l *log.Logger) {
	date := params["date"]
	if date != todayString() {
		ctx.Abort(http.StatusBadRequest, "Past entries are not editable")
		return
	}

	requestBody, err := ioutil.ReadAll(ctx.Request.Body)
	if err != nil {
		ctx.Abort(http.StatusInternalServerError, err.Error())
		return
	}
	entry := newEntry(user, date)
	err = json.Unmarshal(requestBody, &entry)
	if err != nil {
		ctx.Abort(http.StatusInternalServerError, err.Error())
		return
	}

	// TODO: Pass entry itself to upsert.
	err = upsertEntry(db, user, date, entry.Body)
	if err != nil {
		ctx.Abort(http.StatusInternalServerError, err.Error())
		return
	}

	// TODO: Return entry as JSON.
	ren.JSON(200, make(map[string]interface{}))
}

func showDates(ctx *web.Context, ren render.Render, db *mgo.Database, params martini.Params, user *User) {
	now := time.Now()
	date, err := parseDate(params["date"])
	if err != nil {
		ctx.Abort(http.StatusInternalServerError, err.Error())
		return
	}
	dates, err := findEntryDates(db, user, date, now)
	if err != nil {
		ctx.Abort(http.StatusInternalServerError, err.Error())
		return
	}

	data := make(map[string]interface{})
	data["EntryDates"] = dates
	data["Today"] = dateStringOfTime(now)

	prev := beginningOfPreviousMonth(date)
	data["PreviousMonth"] = dateStringOfTime(prev)

	next := beginningOfNextMonth(date)
	if next.UnixNano() <= now.UnixNano() {
		data["NextMonth"] = dateStringOfTime(next)
	}
	ren.JSON(200, data)
}
