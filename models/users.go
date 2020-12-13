package models

import (
	"github.com/vinhut/gapura/helpers"

	"time"
)

const tableName = "users"

type UserDatabase interface {
	Find(string, string, interface{}) error
	FindByUid(string, string, interface{}) error
	Create(User) (bool, error)
	Update() (bool, error)
	Delete(string) (bool, error)
	IncrementPost(string) error
	DecrementPost(string) error
	IncrementLike(string) error
	DecrementLike(string) error
	IncrementFollowing(string) error
	DecrementFollowing(string) error
	IncrementFollower(string) error
	DecrementFollower(string) error
}

type userDatabase struct {
	db helper.DatabaseHelper
}

type User struct {
	Username       string
	Email          string
	Password       string
	Role           string
	Lastlogin      time.Time
	Creationtime   time.Time
	Avatarurl      string
	Active         bool
	Screenname     string
	Location       string
	Protected      bool
	Description    string
	Verified       bool
	Followercount  int
	Followingcount int
	Likecount      int
	Postcount      int
	Updatetime     time.Time
	Uid            string `bson:"_id, omitempty"`
}

func NewUser() User {
	user := User{
		Username:       "example",
		Email:          "example@example.com",
		Password:       "--",
		Role:           "standard",
		Lastlogin:      time.Now(),
		Creationtime:   time.Now(),
		Avatarurl:      "",
		Active:         true,
		Screenname:     "",
		Location:       "",
		Protected:      false,
		Description:    "",
		Verified:       false,
		Followercount:  0,
		Followingcount: 0,
		Likecount:      0,
		Postcount:      0,
		Updatetime:     time.Now(),
	}

	return user
}

func NewUserDatabase(db helper.DatabaseHelper) UserDatabase {
	return &userDatabase{
		db: db,
	}
}

func (userdb *userDatabase) Find(column string, value string, result_user interface{}) error {
	err := userdb.db.Query(tableName, column, value, result_user)
	if err != nil {
		return err
	}

	return nil
}

func (userdb *userDatabase) FindByUid(column string, value string, result_user interface{}) error {
	err := userdb.db.QueryByUid(tableName, column, value, result_user)
	if err != nil {
		return err
	}

	return nil
}

func (userdb *userDatabase) Create(user User) (bool, error) {

	user.Uid = userdb.db.CreateID()

	err := userdb.db.Insert(tableName, user)
	if err != nil {
		return false, err
	}

	return true, nil
}

func (userdb *userDatabase) Update() (bool, error) {
	return false, nil
}

func (userdb *userDatabase) Delete(email string) (bool, error) {
	query := &User{
		Email: email,
	}
	err := userdb.db.Delete(tableName, query)
	if err != nil {
		return false, err
	}

	return true, nil
}

func (userdb *userDatabase) IncrementPost(userid string) error {

	err := userdb.db.Increment(tableName, "_id", userid, "Postcount", 1)
	if err != nil {
		return err
	}

	return nil
}

func (userdb *userDatabase) DecrementPost(userid string) error {

	err := userdb.db.Increment(tableName, "_id", userid, "Postcount", -1)
	if err != nil {
		return err
	}

	return nil
}

func (userdb *userDatabase) IncrementLike(userid string) error {

	err := userdb.db.Increment(tableName, "_id", userid, "Likecount", 1)
	if err != nil {
		return err
	}

	return nil
}

func (userdb *userDatabase) DecrementLike(userid string) error {

	err := userdb.db.Increment(tableName, "_id", userid, "Likecount", -1)
	if err != nil {
		return err
	}

	return nil
}

func (userdb *userDatabase) IncrementFollowing(userid string) error {

	err := userdb.db.Increment(tableName, "_id", userid, "Followingcount", 1)
	if err != nil {
		return err
	}

	return nil
}

func (userdb *userDatabase) DecrementFollowing(userid string) error {

	err := userdb.db.Increment(tableName, "_id", userid, "Followingcount", -1)
	if err != nil {
		return err
	}

	return nil
}

func (userdb *userDatabase) IncrementFollower(userid string) error {

	err := userdb.db.Increment(tableName, "_id", userid, "Followercount", 1)
	if err != nil {
		return err
	}

	return nil
}

func (userdb *userDatabase) DecrementFollower(userid string) error {

	err := userdb.db.Increment(tableName, "_id", userid, "Followercount", -1)
	if err != nil {
		return err
	}

	return nil
}
