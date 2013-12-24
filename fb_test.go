package main

import (
	"github.com/codegangsta/martini-contrib/sessions"
	"github.com/codegangsta/martini-contrib/web"
	"net/http"
	"net/http/httptest"
	"testing"
)

//
// Mock render
//
type mockRender struct {
	status int
	name   string
	v      interface{}
}

func (render *mockRender) JSON(status int, v interface{}) {
	render.status = status
	render.v = v
}

func (render *mockRender) HTML(status int, name string, v interface{}) {
	render.status = status
	render.name = name
	render.v = v
}

func (render *mockRender) Error(status int) {
	render.status = status
}

//
// Mock session
//
type mockSession struct {
	v map[interface{}]interface{}
}

func (session *mockSession) Get(key interface{}) interface{} {
	return session.v[key]
}

func (session *mockSession) Set(key interface{}, val interface{}) {
	session.v[key] = val
}

func (session *mockSession) Delete(key interface{}) {
	delete(session.v, key)
}

func (session *mockSession) AddFlash(value interface{}, vars ...string) {
}

func (session *mockSession) Flashes(vars ...string) []interface{} {
	var flashes []interface{}
	return flashes
}

func (session *mockSession) Options(sessions.Options) {
}

func Test_accessTokenUrl(t *testing.T) {
	fb := &FacebookAuth{AppId: "APP_ID", AppSecret: "APP_SECRET"}
	code := "SOME_CODE"
	redirect := "http://somewhere.org/something"
	expected := "https://graph.facebook.com/oauth/access_token?client_id=APP_ID&client_secret=APP_SECRET&code=SOME_CODE&redirect_uri=http%3A%2F%2Fsomewhere.org%2Fsomething"
	if u := accessTokenUrl(fb, code, redirect); u != expected {
		t.Errorf("Expected %s but got %s", expected, u)
	}
}

func Test_showLogin(t *testing.T) {
	render := &mockRender{}
	expectedStatus := 200
	expectedName := "auth"
	expectedFbUrl := "https://www.facebook.com/dialog/oauth?client_id=&redirect_uri=/auth/callback"
	showLogin(render)
	if status := render.status; status != expectedStatus {
		t.Errorf("Expected to set status %d but got %d", expectedStatus, status)
	}
	if name := render.name; name != expectedName {
		t.Errorf("Expected to set name %s but got %s", expectedName, name)
	}
	if url := render.v.(map[string]interface{})["FacebookUrl"]; url != expectedFbUrl {
		t.Errorf("Expected to set FacebookUrl %s but got %s", expectedFbUrl, url)
	}
}

func Test_logout(t *testing.T) {
	w := httptest.NewRecorder()
	r, err := http.NewRequest("", "", nil)
	if err != nil {
		t.Fatal(err)
	}
	ctx := &web.Context{Request: r, ResponseWriter: w}
	v := make(map[interface{}]interface{})
	v[SESSION_USER_ID_KEY] = "SOME_USER_KEY"
	session := &mockSession{v: v}
	logout(ctx, session)

	if _, ok := v[SESSION_USER_ID_KEY]; ok {
		t.Error("Expected to delete session user ID key but didn't")
	}

	expectedCode := 302
	if w.Code != expectedCode {
		t.Errorf("Expected %d but got %d", w.Code, expectedCode)
	}

	expectedLocation := "/auth"
	if loc := w.HeaderMap["Location"][0]; loc != expectedLocation {
		t.Errorf("Expected %s but got %s", expectedLocation, loc)
	}
}
