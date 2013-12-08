package main

import (
  "fmt"
  "log"
  "regexp"
  "os"
  "os/signal"
  "time"
  "html/template"
  "net/http"
  "unicode/utf8"
  "labix.org/v2/mgo"
  "labix.org/v2/mgo/bson"
  "github.com/joho/godotenv"
)

//
// Entry
//
type Entry struct {
  Id   bson.ObjectId `bson:"_id"`
  Date string        `bson:"date"`
  Body string        `bson:"body"`
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

var templates = prepareTemplates("edit.html", "view.html")

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
var validPath = regexp.MustCompile("^/(edit|save|view)/([0-9]+-[0-9]+-[0-9]+)$")

func makeHandler(fn func(http.ResponseWriter, *http.Request, string, *mgo.Collection), entries *mgo.Collection) http.HandlerFunc {
  return func(w http.ResponseWriter, r *http.Request) {
    m := validPath.FindStringSubmatch(r.URL.Path)
    if m == nil {
      http.NotFound(w, r)
      return
    }
    fn(w, r, m[2], entries)
  }
}

func viewHandler(w http.ResponseWriter, r *http.Request, date string, entries *mgo.Collection) {
  var p Entry
  err := entries.Find(bson.M{"date": date}).One(&p)
  if err != nil {
    http.Redirect(w, r, "/edit/" + date, http.StatusFound)
    return
  }
  renderTemplate(w, "view", &p)
}

func editHandler(w http.ResponseWriter, r *http.Request, date string, entries *mgo.Collection) {
  var p Entry
  err := entries.Find(bson.M{"date": date}).One(&p)
  if err != nil {
    p = Entry{Id: bson.NewObjectId(), Date: date, Body: ""}
  }
  renderTemplate(w, "edit", &p)
}

func saveHandler(w http.ResponseWriter, r *http.Request, date string, entries *mgo.Collection) {
  body := r.FormValue("body")
  _, err := entries.Upsert(bson.M{"date": date}, bson.M{"date": date, "body": body})
  if err != nil {
    http.Error(w, err.Error(), http.StatusInternalServerError)
    return
  }
  http.Redirect(w, r, "/view/" + date, http.StatusFound)
}

func rootHandler(w http.ResponseWriter, r *http.Request) {
  if r.URL.Path == "/" {
    now := time.Now()
    today := fmt.Sprintf("%d-%d-%d", now.Year(), now.Month(), now.Day())
    http.Redirect(w, r, "/view/" + today, http.StatusFound)
  } else {
    http.ServeFile(w, r, "public" + r.URL.Path)
  }
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

  entries := session.DB("morning_pages").C("entries")

  http.HandleFunc("/view/", makeHandler(viewHandler, entries))
  http.HandleFunc("/edit/", makeHandler(editHandler, entries))
  http.HandleFunc("/save/", makeHandler(saveHandler, entries))
  http.HandleFunc("/", rootHandler)

  port := os.Getenv("PORT")
  if port == "" {
    port = "8080"
  }
  log.Println("Listening on " + port)
  http.ListenAndServe(":" + port, nil)
}

