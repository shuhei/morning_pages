package main

import (
	"labix.org/v2/mgo/bson"
	"testing"
	"time"
)

func Test_NewEntry(t *testing.T) {
	user := &User{Id: bson.NewObjectId(), Uid: "mm", Name: "Mitsuru Murakami"}
	date := "2013-12-23"
	entry := NewEntry(user, date)

	if entry.Date != date {
		t.Errorf("Expected %s but got %s", date, entry.Date)
	}
	if entry.UserId != user.Id {
		t.Errorf("Expected %v but got %v", user.Id, entry.UserId)
	}
	if entry.Body != "" {
		t.Errorf("Expected new entry body to be empty but got %s", entry.Body)
	}
}

func Test_daysIn_30(t *testing.T) {
	days := daysIn(4, 2012)
	expected := 30
	if days != expected {
		t.Errorf("Expected %d but got %d", expected, days)
	}
}

func Test_daysIn_31(t *testing.T) {
	days := daysIn(1, 2012)
	expected := 31
	if days != expected {
		t.Errorf("Expected %d but got %d", expected, days)
	}
}

func Test_daysIn_feb(t *testing.T) {
	normal := daysIn(2, 2013)
	expected := 28
	if normal != expected {
		t.Errorf("Expected %d but got %d", expected, normal)
	}

	leap := daysIn(2, 2012)
	expectedLeap := 29
	if leap != expectedLeap {
		t.Errorf("Expected %d but got %d", expectedLeap, leap)
	}
}

func Test_dateString(t *testing.T) {
	expected := "2012-08-08"
	if s := dateString(2012, 8, 8); s != expected {
		t.Errorf("Expected %s but got %s", expected, s)
	}
}

func Test_dateStringOfTime_pad(t *testing.T) {
	d := time.Date(2013, 7, 8, 0, 0, 0, 0, time.Local)
	expected := "2013-07-08"
	if s := dateStringOfTime(d); s != expected {
		t.Errorf("Expected %s but got %s", expected, s)
	}
}

func Test_dateStringOfTime_eod(t *testing.T) {
	d := time.Date(2013, 7, 8, 23, 59, 59, 999, time.Local)
	expected := "2013-07-08"
	if s := dateStringOfTime(d); s != expected {
		t.Errorf("Expected %s but got %s", expected, s)
	}
}

func Test_isValidDate_zeropad(t *testing.T) {
	d := "1981-01-02"
	if !isValidDate(d) {
		t.Errorf("Expected %s to be valid date", d)
	}
}

func Test_isValidDate_nopad(t *testing.T) {
	d := "1981-1-2"
	if isValidDate(d) {
		t.Errorf("Expected %s to be invalid date", d)
	}
}

func Test_isValidDate_notdate(t *testing.T) {
	d := "hello"
	if isValidDate(d) {
		t.Errorf("Expected %s to be invalid date", d)
	}
}

func Test_beginningOfPreviousMonth(t *testing.T) {
	d := time.Date(2013, 7, 8, 23, 59, 59, 999, time.Local)
	expected := time.Date(2013, 6, 1, 0, 0, 0, 0, time.Local)
	if prev := beginningOfPreviousMonth(d); prev != expected {
		t.Errorf("Expected %v but got %v", expected, prev)
	}
}

func Test_beginningOfNextMonth(t *testing.T) {
	d := time.Date(2013, 7, 8, 23, 59, 59, 999, time.Local)
	expected := time.Date(2013, 8, 1, 0, 0, 0, 0, time.Local)
	if next := beginningOfNextMonth(d); next != expected {
		t.Errorf("Expected %v but got %v", expected, next)
	}
}

func Test_parseDate(t *testing.T) {
	tokyo := time.FixedZone("JST", 9*60*60)
	expected := time.Date(2014, 4, 1, 0, 0, 0, 0, tokyo)
	d, err := parseDate("2014-04-01")
	if err != nil {
		t.Errorf("Didn't expect error but got %v", err)
	}
	if d.UnixNano() != expected.UnixNano() {
		t.Errorf("Expected %v but got %v", expected, d)
	}
}
