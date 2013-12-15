package main

import (
	"testing"
	"time"
)

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
