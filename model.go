package main

import (
	"labix.org/v2/mgo"
	"labix.org/v2/mgo/bson"
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

func findEntry(db *mgo.Database, user *User, date string) (*Entry, error) {
	var entry Entry
	err := db.C(EntryCollectionName).Find(bson.M{"user_id": user.Id, "date": date}).One(&entry)
	return &entry, err
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
