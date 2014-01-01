package main

import (
	"fmt"
	"github.com/codegangsta/inject"
	"github.com/codegangsta/martini-contrib/render"
	"github.com/codegangsta/martini-contrib/sessions"
	"github.com/codegangsta/martini-contrib/web"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
)

//
// Mock render.Render
//
type mockRender struct {
	status   int
	name     string
	location string
	v        interface{}
}

func (render *mockRender) JSON(status int, v interface{}) {
	render.status = status
	render.v = v
}

func (render *mockRender) HTML(status int, name string, v interface{}, htmlOpt ...render.HTMLOptions) {
	render.status = status
	render.name = name
	render.v = v
}

func (render *mockRender) Error(status int) {
	render.status = status
}

func (render *mockRender) Redirect(location string, status ...int) {
	if len(status) > 0 {
		render.status = status[0]
	} else {
		render.status = 302
	}
	render.location = location
}

//
// Mock sessions.Session
//
type mockSession struct {
	v       map[interface{}]interface{}
	flashes map[interface{}][]interface{}
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
	key := vars[0]
	session.flashes[key] = append(session.flashes[key], value)
}

func (session *mockSession) Flashes(vars ...string) []interface{} {
	key := vars[0]
	values, ok := session.flashes[key]
	if !ok || len(values) == 0 {
		return nil
	}
	return values
}

func (session *mockSession) Options(sessions.Options) {
}

//
// Mock martini.Context
//
type mockContext struct {
	inject.Injector
}

func (ctx *mockContext) Next() {
}

func (ctx *mockContext) Written() bool {
	return false
}

//
// Mock FacebookAuth
//
type mockFacebookAuth struct {
	token FacebookToken
	user  *FacebookUser
	code  string
}

func (fb *mockFacebookAuth) DialogUrl() string {
	return "DIALOG_URL"
}
func (fb *mockFacebookAuth) AccessTokenUrl(code string) string {
	fb.code = code
	return "ACCESS_TOKEN_URL"
}
func (fb *mockFacebookAuth) MyUrl(token FacebookToken) string {
	return "MY_URL"
}
func (fb *mockFacebookAuth) GetAccessToken(tokenUrl string) (FacebookToken, error) {
	return fb.token, nil
}
func (fb *mockFacebookAuth) GetUserInfo(userUrl string) (*FacebookUser, error) {
	return fb.user, nil
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

func TestFacebookAuth_GetAccessToken_err(t *testing.T) {
	fb := NewFacebookAuth("APP_ID", "APP_SECRET", "http://somewhere.org/something")
	_, err := fb.GetAccessToken("NOWHERE")
	if err == nil {
		t.Error("Expected an error but didn't get one")
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

func TestFacebookAuth_GetUserInfo_err(t *testing.T) {
	fb := NewFacebookAuth("APP_ID", "APP_SECRET", "http://somewhere.org/something")
	_, err := fb.GetUserInfo("NOWHERE")
	if err == nil {
		t.Error("Expected an error but didn't get one")
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
	session := &mockSession{}
	fb := NewFacebookAuth("APP_ID", "APP_SECRET", "http://somewhere.org/something")
	expectedStatus := 200
	expectedName := "auth"
	expectedFbUrl := "https://www.facebook.com/dialog/oauth?client_id=APP_ID&redirect_uri=http%3A%2F%2Fsomewhere.org%2Fsomething"
	showLogin(render, session, fb)
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
	v[SessionUserIdKey] = "SOME_USER_KEY"
	session := &mockSession{v: v}
	logout(ctx, session)

	if _, ok := v[SessionUserIdKey]; ok {
		t.Error("Expected to delete session user ID key but didn't")
	}

	expectedCode := 302
	if w.Code != expectedCode {
		t.Errorf("Expected %d but got %d", expectedCode, w.Code)
	}

	expectedLocation := "/auth"
	if loc := w.HeaderMap["Location"][0]; loc != expectedLocation {
		t.Errorf("Expected %s but got %s", expectedLocation, loc)
	}
}

func Test_getAccessToken(t *testing.T) {
	w := httptest.NewRecorder()
	r, err := http.NewRequest("GET", "/somewhere?code=12345", nil)
	if err != nil {
		t.Fatal(err)
	}

	ctx := &web.Context{Request: r, ResponseWriter: w}
	c := &mockContext{inject.New()}
	fb := &mockFacebookAuth{token: FacebookToken("FB_TOKEN")}
	getAccessToken(ctx, c, fb)

	expectedCode := "12345"
	if fb.code != expectedCode {
		t.Errorf("Expected %s but got %s", expectedCode, fb.code)
	}

	expectedToken := FacebookToken("FB_TOKEN")
	token := c.Get(reflect.TypeOf(FacebookToken(""))).Interface().(FacebookToken)
	if token != expectedToken {
		t.Errorf("Expected %s but got %s", expectedToken, token)
	}
}
