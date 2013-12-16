package main

import (
	"github.com/codegangsta/martini"
	"net/http"
	"net/http/httptest"
	"regexp"
	"testing"
)

func Test_rootHandler(t *testing.T) {
	w := httptest.NewRecorder()
	r, err := http.NewRequest("", "", nil)
	if err != nil {
		t.Fatal(err)
	}
	rootHandler(w, r)

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
	r, err := http.NewRequest("", "", nil)
	if err != nil {
		t.Fatal(err)
	}
	p := martini.Params{"date": "2013-1-1"}
	validateDate(w, r, p)

	badRequest := 400
	if w.Code != badRequest {
		t.Errorf("Expected %d but got %d", badRequest, w.Code)
	}
}

func Test_validateDate_valid(t *testing.T) {
	w := httptest.NewRecorder()
	r, err := http.NewRequest("", "", nil)
	if err != nil {
		t.Fatal(err)
	}
	p := martini.Params{"date": "2013-01-01"}
	validateDate(w, r, p)

	badRequest := 200
	if w.Code != badRequest {
		t.Errorf("Expected %d but got %d", badRequest, w.Code)
	}
}
