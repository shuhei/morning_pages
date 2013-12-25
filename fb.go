package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/codegangsta/martini"
	"github.com/codegangsta/martini-contrib/render"
	"github.com/codegangsta/martini-contrib/sessions"
	"github.com/codegangsta/martini-contrib/web"
	"io/ioutil"
	"labix.org/v2/mgo"
	"log"
	"net/http"
	"net/url"
)

type FacebookToken string

type FacebookAuth interface {
	DialogUrl() string
	AccessTokenUrl(code string) string
	GetAccessToken(code string) (FacebookToken, error)
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

func (fb *facebookAuth) GetAccessToken(code string) (FacebookToken, error) {
	res, err := http.Get(fb.AccessTokenUrl(code))
	if err != nil {
		log.Println("Failed to request access token from Facebook")
		return "", err
	}
	defer res.Body.Close()

	bodyBytes, _ := ioutil.ReadAll(res.Body)
	body := string(bodyBytes)
	if res.StatusCode != 200 {
		log.Println("Failed to get access token", body)
		return "", errors.New("Failed to get access token from Facebook.")
	}

	// Find access token in the response body.
	params, _ := url.ParseQuery(body)
	return FacebookToken(params["access_token"][0]), nil
}

//
// Facebook OAuth handlers
//

func showLogin(r render.Render, fb FacebookAuth) {
	data := make(map[string]interface{})
	data["FacebookUrl"] = fb.DialogUrl()
	r.HTML(200, "auth", data)
}

func logout(ctx *web.Context, session sessions.Session) {
	session.Delete(SESSION_USER_ID_KEY)
	ctx.Redirect(http.StatusFound, "/auth")
}

func getAccessToken(ctx *web.Context, c martini.Context, fb FacebookAuth) {
	// TODO: Handle the case user cancelled logging in.

	// Get access token with the code.
	code := ctx.Request.URL.Query()["code"][0]
	token, err := fb.GetAccessToken(code)
	if err != nil {
		ctx.Abort(http.StatusInternalServerError, err.Error())
		return
	}

	c.Map(token)
}

func getUserInfo(ctx *web.Context, token FacebookToken, c martini.Context) {
	// Get user info with the token.
	myUrl := fmt.Sprintf("https://graph.facebook.com/me?access_token=%s", token)
	res, err := http.Get(myUrl)
	if err != nil {
		log.Println("Failed to request user information from Facebook")
		ctx.Abort(http.StatusInternalServerError, err.Error())
		return
	}
	defer res.Body.Close()

	body, _ := ioutil.ReadAll(res.Body)

	myInfo := make(map[string]interface{})
	err = json.Unmarshal(body, &myInfo)
	if err != nil {
		log.Println("Failed to parse JSON of Facebook user info")
		ctx.Abort(http.StatusInternalServerError, err.Error())
		return
	}

	userInfo := &FacebookUser{
		Id:   myInfo["id"].(string),
		Name: myInfo["name"].(string),
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

	session.Set(SESSION_USER_ID_KEY, user.Id.Hex())

	ctx.Redirect(http.StatusFound, "/")
}
