package main

import (
	"github.com/codegangsta/martini"
	"github.com/codegangsta/martini-contrib/render"
	"github.com/codegangsta/martini-contrib/sessions"
	"github.com/codegangsta/martini-contrib/web"
	"labix.org/v2/mgo"
	"log"
	"net/http"
	"strings"
	"time"
)

const SessionUserIdKey string = "user-id"
const FlashErrorKey string = "error"

//
// Middleware
//

type Mime struct {
	accept    string
	param     string
	extension string
}

func (m *Mime) HTML() bool {
	return m.accepts("html")
}

func (m *Mime) XML() bool {
	return m.accepts("xml")
}

func (m *Mime) JSON() bool {
	return m.accepts("json")
}

func (m *Mime) accepts(name string) bool {
	return m.extension == name || m.param == name || strings.Contains(m.accept, name)
}

func mime(ctx *web.Context, c martini.Context) {
	accept := ctx.Request.Header.Get("Accept")
	param := ctx.Params["format"]

	path := ctx.Request.URL.Path
	components := strings.Split(path, ".")
	extension := ""
	if size := len(components); size > 0 {
		extension = components[size-1]
	}

	mime := &Mime{accept, param, extension}
	c.Map(mime)
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
	today := todayString()
	ctx.Redirect(http.StatusFound, "/entries/"+today)
}

func showEntry(ctx *web.Context, ren render.Render, db *mgo.Database, params martini.Params, mime *Mime, session sessions.Session, user *User) {
	date := params["date"]
	today := todayString()
	entry, err := findEntry(db, user, date)
	if err != nil {
		if date == today && mime.HTML() {
			ctx.Redirect(http.StatusFound, "/entries/"+date+"/edit")
			return
		} else {
			entry = newEntry(user, date)
		}
	}

	if mime.JSON() {
		ren.JSON(200, entry)
	} else {
		data := map[string]interface{}{
			"Entry":       entry,
			"CurrentUser": user,
			"IsEditable":  today == date,
			"Error":       getError(session),
		}
		ren.HTML(200, "view", data)
	}
}

func editEntry(ctx *web.Context, r render.Render, db *mgo.Database, params martini.Params, session sessions.Session, user *User) {
	date := params["date"]
	today := todayString()
	if date != today {
		setError("Past entries are not editable", session)
		ctx.Redirect(http.StatusFound, "/entries/"+date)
		return
	}

	entry, err := findEntry(db, user, date)
	if err != nil {
		entry = newEntry(user, date)
	}

	data := map[string]interface{}{
		"Entry":       entry,
		"CurrentUser": user,
		"Error":       getError(session),
	}
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
