package main

import (
  "fmt"
  "log"
  "regexp"
  "io/ioutil"
  "os"
  "os/signal"
  "strings"
  "time"
  "html/template"
  "net/http"
  "net/url"
  "unicode/utf8"
  "encoding/json"
  "labix.org/v2/mgo"
  "labix.org/v2/mgo/bson"
  "github.com/joho/godotenv"
  "github.com/codegangsta/martini"
  "github.com/codegangsta/martini-contrib/sessions"
  "github.com/codegangsta/martini-contrib/render"
)

//
// Models
//
const ENTRY_COLLECTION_NAME = "entries"
const USER_COLLECTION_NAME = "users"

type Entry struct {
  Id     bson.ObjectId `bson:"_id"`
  Date   string        `bson:"date"`
  Body   string        `bson:"body"`
  UserId bson.ObjectId `bson:"user_id"`
}

type User struct {
  Id    bson.ObjectId `bson:"_id"`
  Uid   string        `bson:"uid"`
  Name  string        `bson:"name"`
}

//
// Template
//
func unsafe(str string) template.HTML {
  return template.HTML(str)
}

var linebreakPattern, _ = regexp.Compile("\r?\n")
func linebreak(str string) string {
  // REVIEW: Should I use []byte instead of string?
  return string(linebreakPattern.ReplaceAll([]byte(str), []byte("<br>")))
}

func charCount(str string) int {
  withoutCr := strings.Replace(str, "\r\n", "\n", -1)
  return utf8.RuneCountInString(withoutCr)
}

//
// Handlers
//
const SESSION_USER_ID_KEY string = "user-id"

func authorize(w http.ResponseWriter, r *http.Request, db *mgo.Database, c martini.Context, session sessions.Session) {
  userId := session.Get(SESSION_USER_ID_KEY)
  if userId == nil {
    log.Println("Unauthorized access")
    http.Redirect(w, r, "/auth", http.StatusFound)
    return
  }

  var user User
  err := db.C(USER_COLLECTION_NAME).FindId(bson.ObjectIdHex(userId.(string))).One(&user)
  if err != nil {
    log.Println("User not found")
    http.Redirect(w, r, "/auth", http.StatusFound)
    return
  }

  c.Map(&user)
}

func validateDate(w http.ResponseWriter, r *http.Request, params martini.Params) {
  date := params["date"]
  matched, err := regexp.MatchString("[0-9]+-[0-9]+-[0-9]+", date)
  if err != nil {
    panic(err)
  }
  if !matched {
    log.Println("Invalid date: %s", date)
    http.Redirect(w, r, "/", http.StatusFound)
    return
  }
}

// TODO: Reduce arguments. Can't I redirect without w and req?
func viewHandler(w http.ResponseWriter, r *http.Request, ren render.Render, db *mgo.Database, params martini.Params, user *User) {
  date := params["date"]
  var entry Entry
  err := db.C(ENTRY_COLLECTION_NAME).Find(bson.M{"user_id": user.Id, "date": date}).One(&entry)
  if err != nil {
    http.Redirect(w, r, "/entries/" + date + "/edit", http.StatusFound)
    return
  }
  data := make(map[string]interface{})
  data["Entry"] = &entry
  data["CurrentUser"] = user
  ren.HTML(200, "view", data)
}

func editHandler(r render.Render, db *mgo.Database, params martini.Params, user *User) {
  date := params["date"]
  var entry Entry
  err := db.C(ENTRY_COLLECTION_NAME).Find(bson.M{"user_id": user.Id, "date": date}).One(&entry)
  if err != nil {
    entry = Entry{Id: bson.NewObjectId(), Date: date, Body: "", UserId: user.Id}
  }
  data := make(map[string]interface{})
  data["Entry"] = &entry
  data["CurrentUser"] = user
  r.HTML(200, "edit", data)
}

func saveHandler(w http.ResponseWriter, r *http.Request, db *mgo.Database, params martini.Params, user *User) {
  date := params["date"]
  query := bson.M{"date": date, "user_id": user.Id}
  entry := bson.M{"date": date, "user_id": user.Id, "body": r.FormValue("body")}
  _, err := db.C(ENTRY_COLLECTION_NAME).Upsert(query, entry)
  if err != nil {
    http.Error(w, err.Error(), http.StatusInternalServerError)
    return
  }
  http.Redirect(w, r, "/entries/" + date, http.StatusFound)
}

func rootHandler(w http.ResponseWriter, r *http.Request) {
  now := time.Now()
  today := fmt.Sprintf("%d-%d-%d", now.Year(), now.Month(), now.Day())
  http.Redirect(w, r, "/entries/" + today, http.StatusFound)
}

func redirectUrl() string {
  host := os.Getenv("FB_REDIRECT_URL")
  return fmt.Sprintf("%s/auth/callback", host)
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

// TODO: Split this looooong function.
func authCallbackHandler(w http.ResponseWriter, r *http.Request, db *mgo.Database, session sessions.Session) {
  // TODO: Handle the case user cancelled logging in.

  appId := os.Getenv("FB_APP_ID")
  appSecret := os.Getenv("FB_APP_SECRET")
  code := r.URL.Query()["code"][0]

  // Get access token with the code.
  tokenUrl := fmt.Sprintf("https://graph.facebook.com/oauth/access_token?client_id=%s&redirect_uri=%s&client_secret=%s&code=%s", appId, redirectUrl(), appSecret, code)
  res, err := http.Get(tokenUrl)
  if err != nil {
    log.Println("Failed to request access token from Facebook")
    http.Error(w, err.Error(), http.StatusInternalServerError)
    return
  }
  defer res.Body.Close()

  resBody, _ := ioutil.ReadAll(res.Body)
  body := string(resBody)
  if res.StatusCode != 200 {
    log.Println("Failed to get access token", body)
    http.Error(w, "Failed to get access token from Facebook.", http.StatusInternalServerError)
    return
  }

  // Find access token in the response body.
  params, _ := url.ParseQuery(body)
  token := params["access_token"][0]

  // Get user info with the token.
  myUrl := fmt.Sprintf("https://graph.facebook.com/me?access_token=%s", token)
  myRes, err := http.Get(myUrl)
  if err != nil {
    log.Println("Failed to request user information from Facebook")
    http.Error(w, err.Error(), http.StatusInternalServerError)
    return 
  }
  defer myRes.Body.Close()

  myResBody, _ := ioutil.ReadAll(myRes.Body)

  myInfo := make(map[string]interface{})
  err = json.Unmarshal(myResBody, &myInfo)

  myId := myInfo["id"].(string)
  myName := myInfo["name"].(string)

  var user User
  err = db.C(USER_COLLECTION_NAME).Find(bson.M{"uid": myId}).One(&user)
  if (err != nil) {
    user = User{Id: bson.NewObjectId(), Uid: myId, Name: myName}
    err = db.C(USER_COLLECTION_NAME).Insert(user)
    if (err != nil) {
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

func prepareRouter(m *martini.ClassicMartini) {
  m.Get("/", authorize, rootHandler)

  m.Get("/auth", authHandler)
  m.Get("/auth/logout", authLogoutHandler)
  m.Get("/auth/callback", authCallbackHandler)

  m.Get("/entries/:date", authorize, validateDate, viewHandler)
  m.Post("/entries/:date", authorize, validateDate, saveHandler)
  m.Get("/entries/:date/edit", authorize, validateDate, editHandler)
}

//
// Execution
//
func cleanupBeforeExit(cleanup func()) {
  terminateChan := make(chan os.Signal, 1)
  signal.Notify(terminateChan, os.Interrupt)
  signal.Notify(terminateChan, os.Kill)
  go func() {
    <-terminateChan
    cleanup()
  }()
}

func main() {
  err := godotenv.Load()
  if err != nil {
    log.Println("Failed to load .env. Maybe on production?")
  }

  m := martini.Classic()

  //
  // Database
  //
  session, err := mgo.Dial(os.Getenv("MONGOHQ_URL"))
  if (err != nil) {
    log.Fatal(err)
  }
  cleanupBeforeExit(func() {
    log.Println("Cleaning up...")
    session.Close()
    os.Exit(1)
  })

  db := session.DB("") // Use database specified in the URL.
  m.Map(db)

  //
  // Session
  //
  store := sessions.NewCookieStore([]byte(os.Getenv("SESSION_KEY")))
  m.Use(sessions.Sessions("default-session", store))

  //
  // Template
  //
  funcMap := template.FuncMap {
    "unsafe": unsafe,
    "linebreak": linebreak,
    "charCount": charCount,
  }
  m.Use(render.Renderer(render.Options{
    Directory: "templates",
    Extensions: []string{".html"},
    Funcs: []template.FuncMap{funcMap},
    Layout: "layout",
  }))

  prepareRouter(m)

  m.Run()
}
