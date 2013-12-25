package main

import (
	"fmt"
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

func TestFacebookAuth_DialogUrl(t *testing.T) {
	fb := NewFacebookAuth("APP_ID", "APP_SECRET", "http://somewhere.org/something")
	expected := "https://www.facebook.com/dialog/oauth?client_id=APP_ID&redirect_uri=http%3A%2F%2Fsomewhere.org%2Fsomething"
	if u := fb.DialogUrl(); u != expected {
		t.Errorf("Expected %s but got %s", expected, u)
	}
}

func TestFacebookAuth_AccessTokenUrl(t *testing.T) {
	fb := NewFacebookAuth("APP_ID", "APP_SECRET", "http://somewhere.org/something")
	code := "SOME_CODE"
	expected := "https://graph.facebook.com/oauth/access_token?client_id=APP_ID&client_secret=APP_SECRET&code=SOME_CODE&redirect_uri=http%3A%2F%2Fsomewhere.org%2Fsomething"
	if u := fb.AccessTokenUrl(code); u != expected {
		t.Errorf("Expected %s but got %s", expected, u)
	}
}

func TestFacebookAuth_MyUrl(t *testing.T) {
	fb := NewFacebookAuth("APP_ID", "APP_SECRET", "http://somewhere.org/something")
	token := FacebookToken("SOME_TOKEN")
	expected := "https://graph.facebook.com/me?access_token=SOME_TOKEN"
	if u := fb.MyUrl(token); u != expected {
		t.Errorf("Expected %s but got %s", expected, u)
	}
}

func TestFacebookAuth_GetAccessToken_ok(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "access_token=HELLO")
	}))
	defer ts.Close()

	fb := NewFacebookAuth("APP_ID", "APP_SECRET", "http://somewhere.org/something")
	token, err := fb.GetAccessToken(ts.URL)
	if err != nil {
		t.Errorf("Got error %s", err.Error())
	}

	actual := (string)(token)
	expected := "HELLO"
	if actual != expected {
		t.Errorf("Expected %s but got %s", expected, actual)
	}
}

func TestFacebookAuth_GetAccessToken_ng(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(400)
		fmt.Fprintln(w, "Error!")
	}))
	defer ts.Close()

	fb := NewFacebookAuth("APP_ID", "APP_SECRET", "http://somewhere.org/something")
	_, err := fb.GetAccessToken(ts.URL)
	if err == nil {
		t.Error("Expected an error but didn't get one")
	}
}

func TestFacebookAuth_GetUserInfo_ok(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, `{"id": "1234567890", "name": "Hello World"}`)
	}))
	defer ts.Close()

	fb := NewFacebookAuth("APP_ID", "APP_SECRET", "http://somewhere.org/something")
	user, err := fb.GetUserInfo(ts.URL)
	if err != nil {
		t.Errorf("Got error %s", err.Error())
	}

	expectedId := "1234567890"
	expectedName := "Hello World"
	if user.Id != expectedId {
		t.Errorf("Expected %s but got %s", expectedId, user.Id)
	}
	if user.Name != expectedName {
		t.Errorf("Expected %s but got %s", expectedName, user.Name)
	}
}

func TestFacebookAuth_GetUserInfo_ng(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(400)
		fmt.Fprintln(w, "Error!")
	}))
	defer ts.Close()

	fb := NewFacebookAuth("APP_ID", "APP_SECRET", "http://somewhere.org/something")
	_, err := fb.GetUserInfo(ts.URL)
	if err == nil {
		t.Error("Expected an error but didn't get one")
	}
}

func Test_showLogin(t *testing.T) {
	render := &mockRender{}
	fb := NewFacebookAuth("APP_ID", "APP_SECRET", "http://somewhere.org/something")
	expectedStatus := 200
	expectedName := "auth"
	expectedFbUrl := "https://www.facebook.com/dialog/oauth?client_id=APP_ID&redirect_uri=http%3A%2F%2Fsomewhere.org%2Fsomething"
	showLogin(render, fb)
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