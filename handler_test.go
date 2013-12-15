package main

import (
	"net/http"
	"net/http/httptest"
	"regexp"
	"testing"
	"time"
)

func Test_dateString_pad(t *testing.T) {
	d := time.Date(2013, 7, 8, 0, 0, 0, 0, time.Local)
	expected := "2013-07-08"
	if s := dateString(d); s != expected {
		t.Errorf("Expected %s but was %s", expected, s)
	}
}

func Test_dateString_eod(t *testing.T) {
	d := time.Date(2013, 7, 8, 23, 59, 59, 999, time.Local)
	expected := "2013-07-08"
	if s := dateString(d); s != expected {
		t.Errorf("Expected %s but was %s", expected, s)
	}
}

func Test_isValidDate(t *testing.T) {
	d := "1981-01-02"
	if !isValidDate(d) {
		t.Errorf("Expected %s to be valid date", d)
	}

	dd := "1981-1-2"
	if isValidDate(dd) {
		t.Errorf("Expected %s to be invalid date", dd)
	}

	ddd := "hello"
	if isValidDate(ddd) {
		t.Errorf("Expected %s to be invalid date", ddd)
	}
}

func Test_rootHandler(t *testing.T) {
	w := httptest.NewRecorder()
	r, err := http.NewRequest("", "", nil)
	if err != nil {
		t.Fatal(err)
	}
	rootHandler(w, r)

	found := 302
	if w.Code != found {
		t.Errorf("Expected %d but was %d", found, w.Code)
	}

	pattern := "/entries/\\d{4}-\\d{2}-\\d{2}"
	loc := w.HeaderMap["Location"][0]
	matched, _ := regexp.MatchString(pattern, loc)
	if !matched {
		t.Errorf("Expected %s to match %s", loc, pattern)
	}
}
