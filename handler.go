package main

import (
	"github.com/codegangsta/martini"
	"github.com/codegangsta/martini-contrib/render"
	"github.com/codegangsta/martini-contrib/sessions"
	"github.com/codegangsta/martini-contrib/web"
	"labix.org/v2/mgo"
	"log"
	"net/http"
	"time"
)

const SessionUserIdKey string = "user-id"
const FlashErrorKey string = "error"

type TemplateData map[string]interface{}

func NewTemplateData() TemplateData {
	return make(TemplateData)
}

//
// Middleware
//

func initTemplateData(c martini.Context) {
	data := NewTemplateData()
	c.Map(data)
}

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

func fetchDateEntries(ctx *web.Context, db *mgo.Database, data TemplateData, user *User) {
	dates, err := findEntryDates(db, user, time.Now())
	if err != nil {
		ctx.Abort(http.StatusInternalServerError, err.Error())
		return
	}
	data["EntryDates"] = dates
}

//
// Flash
//

func getError(session sessions.Session) string {
	if flashes := session.Flashes(FlashErrorKey); flashes != nil {
		return flashes[0].(string)
	} else {
		return ""
	}
}

func setError(message string, session sessions.Session) {
	session.AddFlash(message, FlashErrorKey)
}

//
// Handlers
//

func rootHandler(ctx *web.Context) {
	today, err := todayString()
	if err != nil {
		ctx.Abort(http.StatusInternalServerError, "Failed to load location")
		return
	}
	ctx.Redirect(http.StatusFound, "/entries/"+today)
}

func showEntry(ctx *web.Context, ren render.Render, db *mgo.Database, params martini.Params, session sessions.Session, data TemplateData, user *User) {
	date := params["date"]
	entry, err := findEntry(db, user, date)
	if err != nil {
		ctx.Redirect(http.StatusFound, "/entries/"+date+"/edit")
		return
	}

	today, err := todayString()
	if err != nil {
		ctx.Abort(http.StatusInternalServerError, "Failed to load location")
		return
	}

	data["Entry"] = entry
	data["CurrentUser"] = user
	data["IsEditable"] = today == date
	data["Error"] = getError(session)
	ren.HTML(200, "view", data)
}

func editEntry(ctx *web.Context, r render.Render, db *mgo.Database, params martini.Params, session sessions.Session, data TemplateData, user *User) {
	date := params["date"]
	today, err := todayString()
	if err != nil {
		ctx.Abort(http.StatusInternalServerError, "Failed to load location")
		return
	}
	if date != today {
		setError("Past entries are not editable", session)
		ctx.Redirect(http.StatusFound, "/entries/"+date)
		return
	}

	entry, err := findEntry(db, user, date)
	if err != nil {
		entry = newEntry(user, date)
	}

	data["Entry"] = entry
	data["CurrentUser"] = user
	data["Error"] = getError(session)
	r.HTML(200, "edit", data)
}

func saveEntry(ctx *web.Context, db *mgo.Database, params martini.Params, user *User) {
	date := params["date"]
	body := ctx.Request.FormValue("body")
	err := upsertEntry(db, user, date, body)
	if err != nil {
		ctx.Abort(http.StatusInternalServerError, err.Error())
		return
	}
	ctx.Redirect(http.StatusFound, "/entries/"+date)
}
