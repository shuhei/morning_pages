package main

import (
	"testing"
)

func Test_unsafe(t *testing.T) {
	str := "<p>Hello, World!</p>"
	html := string(unsafe(str))
	if html != str {
		t.Errorf("Expected %s but got %s", str, html)
	}
}

func Test_charCount(t *testing.T) {
	str := "hello\r\n\r\nworld\n!"
	expected := 14
	count := charCount(str)
	if count != expected {
		t.Errorf("Expected %d but got %d", expected, count)
	}
}

func Test_linebreak(t *testing.T) {
	str := "hello\r\n\r\nworld\n!"
	expected := "hello<br><br>world<br>!"
	broken := linebreak(str)
	if broken != expected {
		t.Errorf("Expected %s but got %s", expected, broken)
	}
}

func Test_extractDay_nopad(t *testing.T) {
	date := "2013-04-14"
	expected := "14"
	day := extractDay(date)
	if day != expected {
		t.Errorf("Expected %s but got %s", expected, day)
	}
}

func Test_extractDay_pad(t *testing.T) {
	date := "2013-01-09"
	expected := "9"
	day := extractDay(date)
	if day != expected {
		t.Errorf("Expected %s but got %s", expected, day)
	}
}
