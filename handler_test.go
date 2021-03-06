package main

import (
	"github.com/codegangsta/martini"
	"github.com/codegangsta/martini-contrib/web"
	"net/http/httptest"
	"testing"
)

func Test_ValidateDate_invalid(t *testing.T) {
	w := httptest.NewRecorder()
	ctx := &web.Context{ResponseWriter: w}
	p := martini.Params{"date": "2013-1-1"}
	ValidateDate(ctx, p)

	badRequest := 400
	if w.Code != badRequest {
		t.Errorf("Expected %d but got %d", badRequest, w.Code)
	}
}

func Test_ValidateDate_valid(t *testing.T) {
	w := httptest.NewRecorder()
	ctx := &web.Context{ResponseWriter: w}
	p := martini.Params{"date": "2013-01-01"}
	ValidateDate(ctx, p)

	badRequest := 200
	if w.Code != badRequest {
		t.Errorf("Expected %d but got %d", badRequest, w.Code)
	}
}
