package main

import (
	"encoding/json"
	"errors"
	"github.com/codegangsta/martini"
	"github.com/codegangsta/martini-contrib/render"
	"github.com/codegangsta/martini-contrib/sessions"
	"github.com/codegangsta/martini-contrib/web"
	"io/ioutil"
	"labix.org/v2/mgo"
	"log"
	"net/http"
	"net/url"
	"strings"
)

type FacebookToken string

type FacebookAuth interface {
	DialogUrl() string
	AccessTokenUrl(code string) string
	MyUrl(token FacebookToken) string

	GetAccessToken(tokenUrl string) (FacebookToken, error)
	GetUserInfo(userUrl string) (*FacebookUser, error)
}

type facebookAuth struct {
	AppId       string
	AppSecret   string
	RedirectUrl string
}

func NewFacebookAuth(appId, appSecret, redirectUrl string) FacebookAuth {
	return &facebookAuth{AppId: appId, AppSecret: appSecret, RedirectUrl: redirectUrl}
}

func (fb *facebookAuth) DialogUrl() string {
	baseUrl := "https://www.facebook.com/dialog/oauth?"

	params := url.Values{}
	params.Add("client_id", fb.AppId)
	params.Add("redirect_uri", fb.RedirectUrl)

	return baseUrl + params.Encode()
}

func (fb *facebookAuth) AccessTokenUrl(code string) string {
	baseUrl := "https://graph.facebook.com/oauth/access_token?"

	params := url.Values{}
	params.Add("client_id", fb.AppId)
	params.Add("redirect_uri", fb.RedirectUrl)
	params.Add("client_secret", fb.AppSecret)
	params.Add("code", code)

	return baseUrl + params.Encode()
}

func (fb *facebookAuth) MyUrl(token FacebookToken) string {
	baseUrl := "https://graph.facebook.com/me?"

	params := url.Values{}
	params.Add("access_token", string(token))

	return baseUrl + params.Encode()
}

func (fb *facebookAuth) GetAccessToken(tokenUrl string) (FacebookToken, error) {
	res, err := http.Get(tokenUrl)
	if err != nil {
		return "", err
	}
	defer res.Body.Close()

	bodyBytes, _ := ioutil.ReadAll(res.Body)
	body := string(bodyBytes)
	if res.StatusCode != 200 {
		return "", errors.New("Failed to get access token from Facebook.")
	}

	// Find access token in the response body.
	params, _ := url.ParseQuery(strings.TrimSpace(body))
	return FacebookToken(params["access_token"][0]), nil
}

func (fb *facebookAuth) GetUserInfo(userUrl string) (*FacebookUser, error) {
	res, err := http.Get(userUrl)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	body, _ := ioutil.ReadAll(res.Body)

	myInfo := make(map[string]interface{})
	err = json.Unmarshal(body, &myInfo)
	if err != nil {
		return nil, err
	}

	userInfo := &FacebookUser{
		Id:   myInfo["id"].(string),
		Name: myInfo["name"].(string),
	}
	return userInfo, nil
}

//
// Handlers
//

func showLogin(r render.Render, fb FacebookAuth) {
	data := make(map[string]interface{})
	data["FacebookUrl"] = fb.DialogUrl()
	r.HTML(200, "auth", data)
}

func logout(ctx *web.Context, session sessions.Session) {
	session.Delete(SessionUserIdKey)
	ctx.Redirect(http.StatusFound, "/auth")
}

func getAccessToken(ctx *web.Context, c martini.Context, fb FacebookAuth) {
	// TODO: Handle the case user cancelled logging in.

	// Get access token with the code.
	codes, ok := ctx.Request.URL.Query()["code"]
	if !ok {
		ctx.Abort(http.StatusInternalServerError, "No code is given")
		return
	}
	code := codes[0]
	tokenUrl := fb.AccessTokenUrl(code)
	token, err := fb.GetAccessToken(tokenUrl)
	if err != nil {
		ctx.Abort(http.StatusInternalServerError, err.Error())
		return
	}

	c.Map(token)
}

func getUserInfo(ctx *web.Context, c martini.Context, token FacebookToken, fb FacebookAuth) {
	userUrl := fb.MyUrl(token)
	userInfo, err := fb.GetUserInfo(userUrl)
	if err != nil {
		ctx.Abort(http.StatusInternalServerError, err.Error())
		return
	}

	c.Map(userInfo)
}

func findOrCreateUser(ctx *web.Context, fbUser *FacebookUser, db *mgo.Database, session sessions.Session) {
	user, err := findFacebookUser(db, fbUser)
	if err != nil {
		user, err = insertFacebookUser(db, fbUser)
		if err != nil {
			log.Println("Failed to create a user")
			log.Println(err)
			ctx.Redirect(http.StatusFound, "/auth")
			return
		}
		log.Println("Created a new user", user.Id)
	} else {
		log.Println("Found a user", user.Id)
	}

	session.Set(SessionUserIdKey, user.Id.Hex())

	ctx.Redirect(http.StatusFound, "/")
}
