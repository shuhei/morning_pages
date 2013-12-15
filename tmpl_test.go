package main

import (
	"testing"
)

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
