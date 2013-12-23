package main

import (
	"fmt"
	"labix.org/v2/mgo"
	"labix.org/v2/mgo/bson"
	"time"
)

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

func findUserById(db *mgo.Database, userId string) (*User, error) {
	var user User
	err := db.C(UserCollectionName).FindId(bson.ObjectIdHex(userId)).One(&user)
	return &user, err
}

func findFacebookUser(db *mgo.Database, fbUser *FacebookUser) (*User, error) {
	var user User
	err := db.C(UserCollectionName).Find(bson.M{"uid": fbUser.Id}).One(&user)
	return &user, err
}

func insertFacebookUser(db *mgo.Database, fbUser *FacebookUser) (*User, error) {
	user := &User{Id: bson.NewObjectId(), Uid: fbUser.Id, Name: fbUser.Name}
	err := db.C(UserCollectionName).Insert(user)
	if err != nil {
		return nil, err
	}
	return user, nil
}

const EntryCollectionName = "entries"

type Entry struct {
	Id     bson.ObjectId `bson:"_id"`
	Date   string        `bson:"date"`
	Body   string        `bson:"body"`
	UserId bson.ObjectId `bson:"user_id"`
}

type DateEntry struct {
	Date     string
	HasEntry bool
	IsFuture bool
}

func findEntry(db *mgo.Database, user *User, date string) (*Entry, error) {
	var entry Entry
	err := db.C(EntryCollectionName).Find(bson.M{"user_id": user.Id, "date": date}).One(&entry)
	return &entry, err
}

func findEntryDates(db *mgo.Database, user *User, t time.Time) ([]DateEntry, error) {
	year, month, day := t.Year(), t.Month(), t.Day()
	days := daysIn(month, year)
	start, end := dateString(year, month, 1), dateString(year, month, days)

	query := bson.M{
		"user_id": user.Id,
		"date":    bson.M{"$gte": start, "$lte": end},
	}
	selector := bson.M{"date": 1}
	var entries []Entry
	err := db.C(EntryCollectionName).Find(query).Select(selector).Sort("date").All(&entries)
	if err != nil {
		return nil, err
	}

	dates := make([]DateEntry, days)
	for i := 0; i < days; i++ {
		date := dateString(year, month, i+1)
		hasEntry := false
		for _, entry := range entries {
			if entry.Date == date {
				hasEntry = true
				break
			}
		}
		dates[i] = DateEntry{Date: date, HasEntry: hasEntry, IsFuture: day < i+1}
	}
	return dates, nil
}

func upsertEntry(db *mgo.Database, user *User, date string, body string) error {
	query := bson.M{"date": date, "user_id": user.Id}
	entry := bson.M{"date": date, "user_id": user.Id, "body": body}
	_, err := db.C(EntryCollectionName).Upsert(query, entry)
	return err
}

func newEntry(user *User, date string) *Entry {
	return &Entry{Id: bson.NewObjectId(), Date: date, Body: "", UserId: user.Id}
}

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
	_, err := time.ParseInLocation("2006-01-02", date, time.Local)
	return err == nil
}

func todayString() (string, error) {
	// TODO: Use user's timezone.
	tokyo, err := time.LoadLocation("Asia/Tokyo")
	if err != nil {
		return "", err
	}
	timeInTokyo := time.Now().In(tokyo)
	return dateStringOfTime(timeInTokyo), nil
}
