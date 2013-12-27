package main

import (
	"github.com/codegangsta/martini"
	"github.com/codegangsta/martini-contrib/web"
	"net/http"
	"net/http/httptest"
	"regexp"
	"testing"
)

func Test_getError(t *testing.T) {
	flashes := make(map[interface{}][]interface{})
	flashes[FlashErrorKey] = []interface{}{"Hello!", "World!"}
	session := &mockSession{flashes: flashes}
	expected := "Hello!"
	if e := getError(session); e != expected {
		t.Errorf("Expected %s but got %s", expected, e)
	}
}

func Test_setError(t *testing.T) {
	flashes := make(map[interface{}][]interface{})
	session := &mockSession{flashes: flashes}
	setError("Hello!", session)
	e := flashes[FlashErrorKey][0]
	expected := "Hello!"
	if e != expected {
		t.Errorf("Expected %s but got %s", expected, e)
	}
}

func Test_rootHandler(t *testing.T) {
	w := httptest.NewRecorder()
	r, err := http.NewRequest("", "", nil)
	if err != nil {
		t.Fatal(err)
	}
	ctx := &web.Context{Request: r, ResponseWriter: w}
	rootHandler(ctx)

	found := 302
	if w.Code != found {
		t.Errorf("Expected %d but got %d", found, w.Code)
	}

	pattern := "/entries/\\d{4}-\\d{2}-\\d{2}"
	loc := w.HeaderMap["Location"][0]
	matched, _ := regexp.MatchString(pattern, loc)
	if !matched {
		t.Errorf("Expected %s to match %s", loc, pattern)
	}
}

func Test_validateDate_invalid(t *testing.T) {
	w := httptest.NewRecorder()
	ctx := &web.Context{ResponseWriter: w}
	p := martini.Params{"date": "2013-1-1"}
	validateDate(ctx, p)

	badRequest := 400
	if w.Code != badRequest {
		t.Errorf("Expected %d but got %d", badRequest, w.Code)
	}
}

func Test_validateDate_valid(t *testing.T) {
	w := httptest.NewRecorder()
	ctx := &web.Context{ResponseWriter: w}
	p := martini.Params{"date": "2013-01-01"}
	validateDate(ctx, p)

	badRequest := 200
	if w.Code != badRequest {
		t.Errorf("Expected %d but got %d", badRequest, w.Code)
	}
}
