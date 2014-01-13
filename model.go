package main

import (
	"fmt"
	"labix.org/v2/mgo"
	"labix.org/v2/mgo/bson"
	"time"
)

//
// User
//

const UserCollectionName = "users"

type User struct {
	Id   bson.ObjectId `bson:"_id"`
	Uid  string        `bson:"uid"`
	Name string        `bson:"name"`
}

type FacebookUser struct {
	Id   string
	Name string
}

type UserStore interface {
	Get(userId string) (*User, error)
	FindByFacebook(fbUser *FacebookUser) (*User, error)
	CreateByFacebook(fbUser *FacebookUser) (*User, error)
}

type userStore struct {
	db *mgo.Database
}

func (store *userStore) Get(userId string) (*User, error) {
	var user User
	err := store.db.C(UserCollectionName).FindId(bson.ObjectIdHex(userId)).One(&user)
	return &user, err
}

func (store *userStore) FindByFacebook(fbUser *FacebookUser) (*User, error) {
	var user User
	err := store.db.C(UserCollectionName).Find(bson.M{"uid": fbUser.Id}).One(&user)
	return &user, err
}

func (store *userStore) CreateByFacebook(fbUser *FacebookUser) (*User, error) {
	user := &User{Id: bson.NewObjectId(), Uid: fbUser.Id, Name: fbUser.Name}
	err := store.db.C(UserCollectionName).Insert(user)
	if err != nil {
		return nil, err
	}
	return user, nil
}

//
// Entry
//

const EntryCollectionName = "entries"

type Entry struct {
	Id     bson.ObjectId `bson:"_id" json:"id"`
	Date   string        `bson:"date" json:"date"`
	Body   string        `bson:"body" json:"body"`
	UserId bson.ObjectId `bson:"user_id" json:"userId"`
}

func NewEntry(user *User, date string) *Entry {
	return &Entry{Id: bson.NewObjectId(), Date: date, Body: "", UserId: user.Id}
}

type EntryStore interface {
	Find(user *User, date string) (*Entry, error)
	FindByDate(user *User, from, to string) ([]Entry, error)
	Create(entry *Entry) (bson.ObjectId, error)
	Update(entry *Entry) error
}

type entryStore struct {
	db *mgo.Database
}

func (store *entryStore) Find(user *User, date string) (*Entry, error) {
	var entry Entry
	q := store.db.C(EntryCollectionName).Find(bson.M{"user_id": user.Id, "date": date})
	count, err := q.Count()
	if err != nil {
		return nil, err
	}
	if count == 0 {
		return nil, nil
	}
	err = q.One(&entry)
	return &entry, err
}

func (store *entryStore) FindByDate(user *User, from, to string) ([]Entry, error) {
	var entries []Entry
	dateQuery := bson.M{}
	if from != "" {
		dateQuery["$gte"] = from
	}
	if to != "" {
		dateQuery["$lte"] = to
	}
	query := bson.M{"user_id": user.Id}
	if len(dateQuery) > 0 {
		query["date"] = dateQuery
	}
	fmt.Println(query)
	err := store.db.C(EntryCollectionName).Find(query).Sort("date").All(&entries)
	return entries, err
}

func (store *entryStore) Create(entry *Entry) (bson.ObjectId, error) {
	entry.Id = bson.NewObjectId()
	err := store.db.C(EntryCollectionName).Insert(entry)
	return entry.Id, err
}

func (store *entryStore) Update(entry *Entry) error {
	err := store.db.C(EntryCollectionName).UpdateId(entry.Id, entry)
	return err
}

//
// Utils
//

// https://groups.google.com/forum/#!topic/golang-nuts/W-ezk71hioo
func daysIn(m time.Month, year int) int {
	return time.Date(year, m+1, 0, 0, 0, 0, 0, time.UTC).Day()
}

func dateString(year int, month time.Month, day int) string {
	return fmt.Sprintf("%04d-%02d-%02d", year, month, day)
}

func dateStringOfTime(t time.Time) string {
	return dateString(t.Year(), t.Month(), t.Day())
}

func isValidDate(date string) bool {
	_, err := parseDate(date)
	return err == nil
}

func parseDate(date string) (time.Time, error) {
	// TODO: Use user's timezone.
	tokyo := time.FixedZone("JST", 9*60*60)
	t, err := time.ParseInLocation("2006-01-02", date, tokyo)
	if err != nil {
		return time.Now(), err
	}
	return t, nil
}

func todayString() string {
	// TODO: Use user's timezone.
	tokyo := time.FixedZone("JST", 9*60*60)
	timeInTokyo := time.Now().In(tokyo)
	return dateStringOfTime(timeInTokyo)
}

func beginningOfPreviousMonth(t time.Time) time.Time {
	y, m, _ := t.Date()
	return time.Date(y, m-1, 1, 0, 0, 0, 0, t.Location())
}

func beginningOfNextMonth(t time.Time) time.Time {
	y, m, _ := t.Date()
	return time.Date(y, m+1, 1, 0, 0, 0, 0, t.Location())
}
