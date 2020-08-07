package main

import (
	"github.com/gin-gonic/gin"
	opentracing "github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	jaegercfg "github.com/uber/jaeger-client-go/config"
	jaegerlog "github.com/uber/jaeger-client-go/log"
	"github.com/uber/jaeger-lib/metrics"
	"github.com/vinhut/gapura/helpers"
	"github.com/vinhut/gapura/models"
	"github.com/vinhut/gapura/utils"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"golang.org/x/crypto/bcrypt"

	"crypto/rand"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"time"
)

func setupRouter(userdb models.UserDatabase) *gin.Engine {

	var JAEGER_COLLECTOR_ENDPOINT = os.Getenv("JAEGER_COLLECTOR_ENDPOINT")
	cfg := jaegercfg.Configuration{
		ServiceName: "auth-service",
		Sampler: &jaegercfg.SamplerConfig{
			Type:  "const",
			Param: 1,
		},
		Reporter: &jaegercfg.ReporterConfig{
			LogSpans:          true,
			CollectorEndpoint: JAEGER_COLLECTOR_ENDPOINT,
		},
	}
	jLogger := jaegerlog.StdLogger
	jMetricsFactory := metrics.NullFactory
	tracer, _, _ := cfg.NewTracer(
		jaegercfg.Logger(jLogger),
		jaegercfg.Metrics(jMetricsFactory),
	)
	opentracing.SetGlobalTracer(tracer)

	key := os.Getenv("KEY")
	router := gin.Default()

	router.GET("/ping", func(c *gin.Context) {
		span := tracer.StartSpan("ping")
		c.String(200, "OK")
		span.Finish()
	})

	router.POST("/login", func(c *gin.Context) {
		spanCtx, _ := tracer.Extract(opentracing.HTTPHeaders, opentracing.HTTPHeadersCarrier(c.Request.Header))
		span := tracer.StartSpan("login", ext.RPCServerOption(spanCtx))

		user_email := c.PostForm("email")
		user_pass := c.PostForm("password")
		result := &models.User{}
		err := userdb.Find("email", user_email, result)

		if err == nil {
			err = bcrypt.CompareHashAndPassword([]byte(result.Password), []byte(user_pass))
		}
		if err == nil {
			iv := make([]byte, 12)
			if _, err := io.ReadFull(rand.Reader, iv); err != nil {
				panic(err.Error())
			}
			now := time.Now()
			auth_token := "{\"uid\": \"" + result.Uid.Hex() + "\", \"email\": \"" + user_email + "\", \"role\": \"" + result.Role + "\", \"created\": \"" + now.Format("2006-01-02T15:04:05") + "\"}"
			token := utils.GCM_encrypt(key, auth_token, iv, nil)
			c.String(200, token)
			span.Finish()

		} else {
			c.String(401, "Unauthorized")
			span.Finish()
		}

	})

	router.GET("/user", func(c *gin.Context) {
		spanCtx, _ := tracer.Extract(opentracing.HTTPHeaders, opentracing.HTTPHeadersCarrier(c.Request.Header))
		span := tracer.StartSpan("check user", ext.RPCServerOption(spanCtx))

		service_name, _ := c.GetQuery("service")
		if service_name == "" {
			c.String(401, "Unauthorized")
		}
		token, exist := c.GetQuery("token")
		if exist == false {
			ret_c, err1 := c.Cookie("token")
			if err1 == nil {
				token = ret_c
			} else {
				span.Finish()
				panic(err1.Error())
			}
		}
		splitted := strings.Split(token, "-")
		ret, err := utils.GCM_decrypt(key, splitted[1], splitted[0], nil)
		if err == nil {
			var placeholder map[string]interface{}
			if err := json.Unmarshal([]byte(ret), &placeholder); err != nil {
				panic(err)
				c.String(401, "Unauthorized")
				span.Finish()
			} else {

				result := &models.User{}
				fmt.Println(placeholder["uid"].(string))
				find_uid, _ := primitive.ObjectIDFromHex(placeholder["uid"].(string))
				err := userdb.FindByUid("_id", find_uid, result)
				if err != nil {
					fmt.Println("error find user")
					span.Finish()
					panic(err)
				}
				user_detail := "{\"uid\": \"" + result.Uid.Hex() +
					"\", \"username\": \"" + result.Username +
					"\", \"email\": \"" + result.Email +
					"\", \"role\": \"" + result.Role +
					"\", \"avatarurl\": \"" + result.Avatarurl +
					"\", \"active\": \"" + strconv.FormatBool(result.Active) +
					"\", \"screenname\": \"" + result.Screenname +
					"\", \"location\": \"" + result.Location +
					"\", \"protected\": \"" + strconv.FormatBool(result.Protected) +
					"\", \"description\": \"" + result.Description +
					"\", \"verified\": \"" + strconv.FormatBool(result.Verified) +
					"\"}"

				c.String(200, user_detail)
				span.Finish()
			}
		} else {
			c.String(401, "Unauthorized")
			span.Finish()
		}
		// TODO : token expiry check
	})

	router.GET("/public", func(c *gin.Context) {
		c.String(200, "Public")
	})

	router.POST("/user", func(c *gin.Context) {
		spanCtx, _ := tracer.Extract(opentracing.HTTPHeaders, opentracing.HTTPHeadersCarrier(c.Request.Header))
		span := tracer.StartSpan("create user", ext.RPCServerOption(spanCtx))
		//service_name := c.PostForm("service")
		user_email := c.PostForm("email")
		password := c.PostForm("password")

		hashed, err := bcrypt.GenerateFromPassword([]byte(password), 14)
		if err != nil {
			fmt.Println(err)
		}

		new_user := &models.User{
			Username:       user_email,
			Email:          user_email,
			Password:       string(hashed),
			Role:           "standard",
			Lastlogin:      time.Now(),
			Creationtime:   time.Now(),
			Avatarurl:      "http://localhost/profile.png",
			Active:         true,
			Screenname:     user_email,
			Location:       "Earth",
			Protected:      false,
			Description:    "Hi please update your profile",
			Verified:       false,
			Followercount:  0,
			Followingcount: 0,
			Likecount:      0,
			Postcount:      0,
			Updatetime:     time.Now(),
			Uid:            primitive.NewObjectIDFromTimestamp(time.Now()),
		}

		_, err = userdb.Create(new_user)

		if err == nil {
			c.String(200, "created")
			span.Finish()
		} else {
			c.String(503, "fail")
			fmt.Println(err)
			span.Finish()
		}

	})

	return router

}

func main() {

	mongo_layer := helper.NewMongoDatabase()
	userdb := models.NewUserDatabase(mongo_layer)
	router := setupRouter(userdb)
	router.Run(":8080") // listen and serve on 0.0.0.0:8080

}
