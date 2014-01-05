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
	"strings"
	"time"
)

const SessionUserIdKey string = "user-id"

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
// Handlers
//

func rootHandler(ctx *web.Context) {
	today := todayString()
	ctx.Redirect(http.StatusFound, "/entries/"+today)
}

func showEntry(ctx *web.Context, ren render.Render, db *mgo.Database, params martini.Params, mime *Mime, user *User) {
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

	if mime.HTML() {
		data := map[string]interface{}{
			"CurrentUser": user,
		}
		ren.HTML(200, "view", data)
	} else {
		ren.JSON(200, entry)
	}
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
