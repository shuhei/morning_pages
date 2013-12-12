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
  "github.com/gorilla/context"
  "github.com/gorilla/mux"
  "github.com/gorilla/sessions"
)

//
// Models
//
type Entry struct {
  Id     bson.ObjectId `bson:"_id"`
  Date   string        `bson:"date"`
  Body   string        `bson:"body"`
  UserId bson.ObjectId `bson:"user_id"`
}

var entries *mgo.Collection

type User struct {
  Id    bson.ObjectId `bson:"_id"`
  Uid   string        `bson:"uid"`
  Name  string        `bson:"name"`
}

var users *mgo.Collection

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

type TemplateMap map[string]*template.Template

func prepareTemplates(filenames ...string) TemplateMap {
  funcMap := template.FuncMap {
    "unsafe": unsafe,
    "linebreak": linebreak,
    "charCount": charCount,
  }
  tmpls := make(TemplateMap)
  for _, filename := range filenames {
    files := []string{"views/" + filename, "views/layout.html"}
    tmpls[filename] = template.Must(template.New("").Funcs(funcMap).ParseFiles(files...))
  }
  return tmpls
}

var templates = prepareTemplates("edit.html", "view.html", "auth.html")

func renderTemplate(w http.ResponseWriter, tmpl string, data interface{}) {
  err := templates[tmpl + ".html"].ExecuteTemplate(w, "layout", data)
  if err != nil {
    http.Error(w, err.Error(), http.StatusInternalServerError)
    return
  }
}

//
// Context
//
const CONTEXT_CURRENT_USER_KEY string = "current-user"

func GetCurrentUser(r *http.Request) *User {
  if user := context.Get(r, CONTEXT_CURRENT_USER_KEY); user != nil {
    return user.(*User)
  }
  return nil
}

func SetCurrentUser(r *http.Request, user *User) {
  context.Set(r, CONTEXT_CURRENT_USER_KEY, user)
}

//
// Handlers
//
var sessionStore = sessions.NewCookieStore([]byte(os.Getenv("SESSION_KEY")))
const SESSION_NAME string = "default-session"
const SESSION_USER_ID_KEY string = "user-id"

func makeHandler(fn func(http.ResponseWriter, *http.Request, string)) http.HandlerFunc {
  return func(w http.ResponseWriter, r *http.Request) {
    // Authenticate
    session, _ := sessionStore.Get(r, SESSION_NAME)
    userId := session.Values[SESSION_USER_ID_KEY]
    if userId == nil {
      log.Println("Unauthorized access")
      http.Redirect(w, r, "/auth", http.StatusFound)
      return
    }
    var user User
    err := users.FindId(bson.ObjectIdHex(userId.(string))).One(&user)
    if err != nil {
      log.Println("User not found")
      http.Redirect(w, r, "/auth", http.StatusFound)
      return
    }
    SetCurrentUser(r, &user)

    date := mux.Vars(r)["date"]
    fn(w, r, date)
  }
}

func viewHandler(w http.ResponseWriter, r *http.Request, date string) {
  user := GetCurrentUser(r)
  var entry Entry
  err := entries.Find(bson.M{"user_id": user.Id, "date": date}).One(&entry)
  if err != nil {
    http.Redirect(w, r, "/entries/" + date + "/edit", http.StatusFound)
    return
  }
  data := make(map[string]interface{})
  data["Entry"] = &entry
  data["CurrentUser"] = GetCurrentUser(r)
  renderTemplate(w, "view", data)
}

func editHandler(w http.ResponseWriter, r *http.Request, date string) {
  user := GetCurrentUser(r)
  var entry Entry
  err := entries.Find(bson.M{"user_id": user.Id, "date": date}).One(&entry)
  if err != nil {
    entry = Entry{Id: bson.NewObjectId(), Date: date, Body: "", UserId: user.Id}
  }
  data := make(map[string]interface{})
  data["Entry"] = &entry
  data["CurrentUser"] = GetCurrentUser(r)
  renderTemplate(w, "edit", data)
}

func saveHandler(w http.ResponseWriter, r *http.Request, date string) {
  body := r.FormValue("body")
  user := GetCurrentUser(r)
  _, err := entries.Upsert(bson.M{"date": date, "user_id": user.Id}, bson.M{"date": date, "user_id": user.Id, "body": body})
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

func staticHandler(w http.ResponseWriter, r *http.Request) {
  http.ServeFile(w, r, "public" + r.URL.Path)
}

func redirectUrl() string {
  host := os.Getenv("FB_REDIRECT_URL")
  return fmt.Sprintf("%s/auth/callback", host)
}

func authHandler(w http.ResponseWriter, r *http.Request) {
  appId := os.Getenv("FB_APP_ID")
  dialogUrl := fmt.Sprintf("https://www.facebook.com/dialog/oauth?client_id=%s&redirect_uri=%s", appId, redirectUrl())
  data := make(map[string]interface{})
  data["FacebookUrl"] = dialogUrl
  err := templates["auth.html"].ExecuteTemplate(w, "layout", data)
  if err != nil {
    http.Error(w, err.Error(), http.StatusInternalServerError)
    return
  }
}

func authLogoutHandler(w http.ResponseWriter, r *http.Request) {
  session, _ := sessionStore.Get(r, SESSION_NAME)
  session.Values[SESSION_USER_ID_KEY] = nil
  session.Save(r, w)
  http.Redirect(w, r, "/auth", http.StatusFound)
}

// TODO: Split this looooong function.
func authCallbackHandler(w http.ResponseWriter, r *http.Request) {
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
  err = users.Find(bson.M{"uid": myId}).One(&user)
  if (err != nil) {
    user = User{Id: bson.NewObjectId(), Uid: myId, Name: myName}
    err = users.Insert(user)
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

  // Save user id in session.
  session, _ := sessionStore.Get(r, SESSION_NAME)
  session.Values[SESSION_USER_ID_KEY] = user.Id.Hex()
  session.Save(r, w)

  http.Redirect(w, r, "/", http.StatusFound)
}

func prepareRouter() *mux.Router {
  r := mux.NewRouter()
  datePattern := "{date:[0-9]+-[0-9]+-[0-9]+}"
  r.HandleFunc("/", rootHandler).Methods("GET")
  r.HandleFunc("/auth", authHandler)
  r.HandleFunc("/auth/logout", authLogoutHandler)
  r.HandleFunc("/auth/callback", authCallbackHandler)
  r.HandleFunc("/entries/" + datePattern, makeHandler(viewHandler)).Methods("GET")
  r.HandleFunc("/entries/" + datePattern, makeHandler(saveHandler)).Methods("POST")
  r.HandleFunc("/entries/" + datePattern + "/edit", makeHandler(editHandler)).Methods("GET")
  r.HandleFunc("/{filepath:.+}", staticHandler).Methods("GET")
  return r
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
  entries = db.C("entries")
  users = db.C("users")

  router := prepareRouter()

  port := os.Getenv("PORT")
  if port == "" {
    port = "5000"
  }
  log.Println("Listening on " + port)
  http.ListenAndServe(":" + port, router)
}
