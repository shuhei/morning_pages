package main

import (
	"encoding/json"
	"fmt"
	"github.com/codegangsta/martini"
	"github.com/codegangsta/martini-contrib/render"
	"github.com/codegangsta/martini-contrib/sessions"
	"io/ioutil"
	"labix.org/v2/mgo"
	"labix.org/v2/mgo/bson"
	"log"
	"net/http"
	"net/url"
	"os"
)

type FacebookAuth struct {
	AppId     string
	AppSecret string
}

type FacebookToken string

type FacebookUser struct {
	Id   string
	Name string
}

func redirectUrl() string {
	host := os.Getenv("FB_REDIRECT_URL")
	return fmt.Sprintf("%s/auth/callback", host)
}

func accessTokenUrl(fb *FacebookAuth, code string) string {
	baseUrl := "https://graph.facebook.com/oauth/access_token?"

	params := url.Values{}
	params.Add("client_id", fb.AppId)
	params.Add("redirect_uri", redirectUrl())
	params.Add("client_secret", fb.AppSecret)
	params.Add("code", code)

	return baseUrl + params.Encode()
}

func authHandler(r render.Render) {
	appId := os.Getenv("FB_APP_ID")
	dialogUrl := fmt.Sprintf("https://www.facebook.com/dialog/oauth?client_id=%s&redirect_uri=%s", appId, redirectUrl())
	data := make(map[string]interface{})
	data["FacebookUrl"] = dialogUrl
	r.HTML(200, "auth", data)
}

func authLogoutHandler(w http.ResponseWriter, r *http.Request, session sessions.Session) {
	session.Set(SESSION_USER_ID_KEY, nil)
	http.Redirect(w, r, "/auth", http.StatusFound)
}

func getAccessToken(w http.ResponseWriter, r *http.Request, c martini.Context, fb *FacebookAuth) {
	// TODO: Handle the case user cancelled logging in.

	// Get access token with the code.
	code := r.URL.Query()["code"][0]
	res, err := http.Get(accessTokenUrl(fb, code))
	if err != nil {
		log.Println("Failed to request access token from Facebook")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer res.Body.Close()

	bodyBytes, _ := ioutil.ReadAll(res.Body)
	body := string(bodyBytes)
	if res.StatusCode != 200 {
		log.Println("Failed to get access token", body)
		http.Error(w, "Failed to get access token from Facebook.", http.StatusInternalServerError)
		return
	}

	// Find access token in the response body.
	params, _ := url.ParseQuery(body)
	token := FacebookToken(params["access_token"][0])
	c.Map(token)
}

func getUserInfo(w http.ResponseWriter, token FacebookToken, c martini.Context) {
	// Get user info with the token.
	myUrl := fmt.Sprintf("https://graph.facebook.com/me?access_token=%s", token)
	res, err := http.Get(myUrl)
	if err != nil {
		log.Println("Failed to request user information from Facebook")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer res.Body.Close()

	body, _ := ioutil.ReadAll(res.Body)

	myInfo := make(map[string]interface{})
	err = json.Unmarshal(body, &myInfo)
	if err != nil {
		log.Println("Failed to parse JSON of Facebook user info")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	userInfo := &FacebookUser{
		Id:   myInfo["id"].(string),
		Name: myInfo["name"].(string),
	}
	c.Map(userInfo)
}

func findOrCreateUser(w http.ResponseWriter, r *http.Request, fbUser *FacebookUser, db *mgo.Database, session sessions.Session) {
	var user User
	err := db.C(USER_COLLECTION_NAME).Find(bson.M{"uid": fbUser.Id}).One(&user)
	if err != nil {
		user = User{Id: bson.NewObjectId(), Uid: fbUser.Id, Name: fbUser.Name}
		err = db.C(USER_COLLECTION_NAME).Insert(user)
		if err != nil {
			log.Println("Failed to create a user")
			log.Println(err)
			http.Redirect(w, r, "/auth", http.StatusFound)
			return
		}
		log.Println("Created a new user", user.Id)
	} else {
		log.Println("Found a user", user.Id)
	}

	session.Set(SESSION_USER_ID_KEY, user.Id.Hex())

	http.Redirect(w, r, "/", http.StatusFound)
}
