package main

import (
  "fmt"
  "log"
  "regexp"
  "io/ioutil"
  "os"
  "os/signal"
  "time"
  "html/template"
  "net/http"
  "net/url"
  "unicode/utf8"
  "encoding/json"
  "labix.org/v2/mgo"
  "labix.org/v2/mgo/bson"
  "github.com/joho/godotenv"
  "github.com/gorilla/mux"
)

//
// Models
//
type Entry struct {
  Id   bson.ObjectId `bson:"_id"`
  Date string        `bson:"date"`
  Body string        `bson:"body"`
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

type TemplateMap map[string]*template.Template

func prepareTemplates(filenames ...string) TemplateMap {
  funcMap := template.FuncMap {
    "unsafe": unsafe,
    "linebreak": linebreak,
    "charCount": utf8.RuneCountInString,
  }
  tmpls := make(TemplateMap)
  for _, filename := range filenames {
    files := []string{"views/" + filename, "views/layout.html"}
    tmpls[filename] = template.Must(template.New("").Funcs(funcMap).ParseFiles(files...))
  }
  return tmpls
}

var templates = prepareTemplates("edit.html", "view.html", "auth.html")

func renderTemplate(w http.ResponseWriter, tmpl string, p *Entry) {
  err := templates[tmpl + ".html"].ExecuteTemplate(w, "layout", p)
  if err != nil {
    http.Error(w, err.Error(), http.StatusInternalServerError)
    return
  }
}

//
// Handlers
//
func makeHandler(fn func(http.ResponseWriter, *http.Request, string)) http.HandlerFunc {
  return func(w http.ResponseWriter, r *http.Request) {
    date := mux.Vars(r)["date"]
    fn(w, r, date)
  }
}

func viewHandler(w http.ResponseWriter, r *http.Request, date string) {
  var p Entry
  err := entries.Find(bson.M{"date": date}).One(&p)
  if err != nil {
    http.Redirect(w, r, "/entries/" + date + "/edit", http.StatusFound)
    return
  }
  renderTemplate(w, "view", &p)
}

func editHandler(w http.ResponseWriter, r *http.Request, date string) {
  var p Entry
  err := entries.Find(bson.M{"date": date}).One(&p)
  if err != nil {
    p = Entry{Id: bson.NewObjectId(), Date: date, Body: ""}
  }
  renderTemplate(w, "edit", &p)
}

func saveHandler(w http.ResponseWriter, r *http.Request, date string) {
  body := r.FormValue("body")
  _, err := entries.Upsert(bson.M{"date": date}, bson.M{"date": date, "body": body})
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

// TODO: Determine redirect URL using the actual hostname and port.
var redirectUrl = "http://localhost:5000/auth/callback"

func authHandler(w http.ResponseWriter, r *http.Request) {
  appId := os.Getenv("FB_APP_ID")
  dialogUrl := fmt.Sprintf("https://www.facebook.com/dialog/oauth?client_id=%s&redirect_uri=%s", appId, redirectUrl)
  err := templates["auth.html"].ExecuteTemplate(w, "layout", dialogUrl)
  if err != nil {
    http.Error(w, err.Error(), http.StatusInternalServerError)
    return
  }
}

func authCallbackHandler(w http.ResponseWriter, r *http.Request) {
  // TODO: Handle the case user cancelled logging in.

  appId := os.Getenv("FB_APP_ID")
  appSecret := os.Getenv("FB_APP_SECRET")
  code := r.URL.Query()["code"][0]

  // Get access token with the code.
  tokenUrl := fmt.Sprintf("https://graph.facebook.com/oauth/access_token?client_id=%s&redirect_uri=%s&client_secret=%s&code=%s", appId, redirectUrl, appSecret, code)
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
    _ = users.Insert(user)
    log.Println("Created a new user", user.Id)
  } else {
    log.Println("Found a user", user.Id)
  }

  // TODO: Save user id in session.

  http.Redirect(w, r, "/", http.StatusFound)
}

func prepareRouter() *mux.Router {
  r := mux.NewRouter()
  datePattern := "{date:[0-9]+-[0-9]+-[0-9]+}"
  r.HandleFunc("/", rootHandler).Methods("GET")
  r.HandleFunc("/auth", authHandler)
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
    log.Fatal(err)
  }

  session, err := mgo.Dial(os.Getenv("MONGO_HOST"))
  if (err != nil) {
    log.Fatal(err)
  }
  cleanupBeforeExit(func() {
    log.Println("Cleaning up...")
    session.Close()
    os.Exit(1)
  })

  db := session.DB("morning_pages")
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
