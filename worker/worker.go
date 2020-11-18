package main

import (
	"github.com/vinhut/gapura/helpers"
	"github.com/vinhut/gapura/models"
	"github.com/vinhut/gapura/services"

	"strings"
)

func increment_post_count(user models.UsersDatabase, msg string) {
	userid := strings.split(msg, " ")[1]
	incr_err := user.IncrementPost(userid)
	if incr_err != nil {
		panic(incr_err)
	}
}

func increment_like_count(user models.UsersDatabase, msg string) {

}

func increment_follower_count(user models.UsersDatabase, msg string) {

}

func increment_following_count(user models.UsersDatabase, msg string) {

}

func decrement_post_count(user models.UsersDatabase, msg string) {

}

func decrement_like_count(user models.UsersDatabase, msg string) {

}

func decrement_follower_count(user models.UsersDatabase, msg string) {

}

func decrement_following_count(user models.UsersDatabase, msg string) {

}

func main() {

	mongo_layer := helper.NewMongoDatabase()
	kafka_service := services.NewKafkaReaderService()
	userdb := models.NewUserDatabase(mongo_layer)

	for {
		msg, read_err := kafka_service.Read()
		if read_err != nil {
			panic(read_err)
		}
		switch true {
		case strings.Contains(msg, "post_add"):
			go increment_post_count(userdb, msg)
		case strings.Contains(msg, "like_add"):
			go increment_like_count(userdb, msg)
		case strings.Contains(msg, "follower_add"):
			go increment_follower_count(userdb, msg)
		case strings.Contains(msg, "following_add"):
			go increment_following_count(userdb, msg)
		case strings.Contains(msg, "post_del"):
			go decrement_post_count(userdb, msg)
		case strings.Contains(msg, "like_del"):
			go decrement_like_count(userdb, msg)
		case strings.Contains(msg, "follower_del"):
			go decrement_follower_count(userdb, msg)
		case strings.Contains(msg, "following_del"):
			go decrement_following_count(userdb, msg)

		}
	}
}
