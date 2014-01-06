package main

import (
	"encoding/json"
	"github.com/codegangsta/martini"
	"github.com/codegangsta/martini-contrib/render"
	"github.com/codegangsta/martini-contrib/sessions"
	"github.com/codegangsta/martini-contrib/web"
	"io/ioutil"
	"log"
	"net/http"
	"time"
)

const SessionUserIdKey string = "user-id"

//
// Filters
//

func Authorize(ctx *web.Context, users UserStore, c martini.Context, session sessions.Session, l *log.Logger) {
	userId := session.Get(SessionUserIdKey)
	if userId == nil || userId == "" {
		l.Println("Unauthorized access")
		ctx.Redirect(http.StatusFound, "/auth")
		return
	}

	user, err := users.Get(userId.(string))
	if err != nil {
		l.Println("User not found")
		session.Delete(SessionUserIdKey)
		ctx.Redirect(http.StatusFound, "/auth")
		return
	}

	c.Map(user)
}

func ValidateDate(ctx *web.Context, params martini.Params) {
	date := params["date"]
	if !isValidDate(date) {
		ctx.Abort(http.StatusBadRequest, "Invalid date. e.g. 2014-01-02")
		return
	}
}

//
// Handlers
//

func ShowRoot(ren render.Render, user *User) {
	data := make(map[string]interface{})
	data["CurrentUser"] = user
	ren.HTML(200, "view", data)
}

//
// JSON APIs
//

func GetEntry(ren render.Render, entries EntryStore, params martini.Params, user *User) {
	date := params["date"]
	entry, err := entries.Find(user, date)
	if err != nil {
		entry = NewEntry(user, date)
	}
	ren.JSON(200, entry)
}

func SaveEntry(ctx *web.Context, ren render.Render, entries EntryStore, params martini.Params, user *User, l *log.Logger) {
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
	entry := NewEntry(user, date)
	err = json.Unmarshal(requestBody, &entry)
	if err != nil {
		ctx.Abort(http.StatusInternalServerError, err.Error())
		return
	}

	// TODO: Pass entry itself to upsert.
	err = entries.Upsert(user, date, entry.Body)
	if err != nil {
		ctx.Abort(http.StatusInternalServerError, err.Error())
		return
	}

	// TODO: Return entry as JSON.
	ren.JSON(200, make(map[string]interface{}))
}

func GetEntryDates(ctx *web.Context, ren render.Render, entries EntryStore, params martini.Params, user *User) {
	now := time.Now()
	date, err := parseDate(params["date"])
	if err != nil {
		ctx.Abort(http.StatusInternalServerError, err.Error())
		return
	}
	dates, err := entries.FindDates(user, date, now)
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
