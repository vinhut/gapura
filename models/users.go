package models

import (
	"github.com/vinhut/gapura/helpers"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"time"
)

const tableName = "users"

type UserDatabase interface {
	Find(string, string, interface{}) error
	Create(*User) (bool, error)
	Update() (bool, error)
	Delete(string) (bool, error)
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
	Verfied        bool
	Followercount  int
	Followingcount int
	Likecount      int
	Postcount      int
	Updatetime     time.Time
	Uid            primitive.ObjectID `bson:"_id, omitempty"`
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
		Verfied:        false,
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

func (userdb *userDatabase) Create(user *User) (bool, error) {
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
