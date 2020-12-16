package main

import (
	"github.com/gin-gonic/gin"
	opentracing "github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	jaeger "github.com/uber/jaeger-client-go"
	jaegercfg "github.com/uber/jaeger-client-go/config"
	jaegerlog "github.com/uber/jaeger-client-go/log"
	transport "github.com/uber/jaeger-client-go/transport/zipkin"
	"github.com/uber/jaeger-client-go/zipkin"
	"github.com/uber/jaeger-lib/metrics"
	helper "github.com/vinhut/gapura/helpers"
	"github.com/vinhut/gapura/models"
	services "github.com/vinhut/gapura/services"
	"github.com/vinhut/gapura/utils"

	"crypto/rand"
	"encoding/json"
	"io"
	"os"
	"strconv"
	"strings"
	"time"
)

var SERVICE_NAME = "auth-service"

func setupRouter(userdb models.UserDatabase, cache services.RedisService) *gin.Engine {

	var JAEGER_COLLECTOR_ENDPOINT = os.Getenv("JAEGER_COLLECTOR_ENDPOINT")
	zipkinPropagator := zipkin.NewZipkinB3HTTPHeaderPropagator()
	trsport, _ := transport.NewHTTPTransport(
		JAEGER_COLLECTOR_ENDPOINT,
		transport.HTTPLogger(jaeger.StdLogger),
	)
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
	cfg.InitGlobalTracer(
		"auth-service",
		jaegercfg.Logger(jLogger),
		jaegercfg.Metrics(jMetricsFactory),
		jaegercfg.Injector(opentracing.HTTPHeaders, zipkinPropagator),
		jaegercfg.Extractor(opentracing.HTTPHeaders, zipkinPropagator),
		jaegercfg.ZipkinSharedRPCSpan(true),
		jaegercfg.Reporter(jaeger.NewRemoteReporter(trsport)),
	)
	tracer := opentracing.GlobalTracer()

	key := os.Getenv("KEY")

	router := gin.Default()

	router.GET("/ping", func(c *gin.Context) {
		c.String(200, "OK")
	})

	router.POST("/login", func(c *gin.Context) {
		spanCtx, _ := tracer.Extract(opentracing.HTTPHeaders, opentracing.HTTPHeadersCarrier(c.Request.Header))
		span := tracer.StartSpan("login", ext.RPCServerOption(spanCtx))

		user_email := c.PostForm("email")
		user_pass := c.PostForm("password")
		result := &models.User{}

		cspan := tracer.StartSpan("find user by email",
			opentracing.ChildOf(span.Context()),
		)
		find_err := userdb.Find("email", user_email, result)
		cspan.Finish()

		if find_err != nil {
			c.AbortWithStatusJSON(401, gin.H{"reason": "login error"})
			return
		}
		cspan = tracer.StartSpan("compare hash",
			opentracing.ChildOf(span.Context()),
		)
		compare_err := utils.CompareHashAndPassword(result.Password, user_pass)
		cspan.Finish()
		if compare_err != nil {
			c.AbortWithStatusJSON(401, gin.H{"reason": "login error"})
			return
		}
		iv := make([]byte, 12)
		if _, read_err := io.ReadFull(rand.Reader, iv); read_err != nil {
			panic(read_err.Error())
		}
		now := time.Now()
		auth_token := "{\"uid\": \"" + result.Uid.Hex() + "\", \"email\": \"" +
			user_email + "\", \"role\": \"" + result.Role +
			"\", \"created\": \"" + now.Format("2006-01-02T15:04:05") +
			"\"}"
		cspan = tracer.StartSpan("gcm encrypt auth_token",
			opentracing.ChildOf(span.Context()),
		)
		token, encrypt_err := utils.GCM_encrypt(key, auth_token, iv, nil)
		cspan.Finish()
		if encrypt_err != nil {
			panic(encrypt_err.Error())
		}
		c.String(200, token)

		span.Finish()
	})

	router.GET("/user", func(c *gin.Context) {
		spanCtx, _ := tracer.Extract(opentracing.HTTPHeaders, opentracing.HTTPHeadersCarrier(c.Request.Header))
		span := tracer.StartSpan("check user", ext.RPCServerOption(spanCtx))

		service_name, _ := c.GetQuery("service")
		if service_name == "" {
			c.AbortWithStatusJSON(401, gin.H{"reason": "Missing service name"})
			return
		}
		token, exist := c.GetQuery("token")
		if exist == false {
			token_cookie, cookie_err := c.Cookie("token")
			if cookie_err != nil {
				span.Finish()
				c.AbortWithStatusJSON(401, gin.H{"reason": "Token not found"})
				return
			} else {
				token = token_cookie
			}
		}
		token_encrypted := strings.Split(token, "-")

		cspan := tracer.StartSpan("gcm decrypt auth_token",
			opentracing.ChildOf(span.Context()),
		)
		token_data, decrypt_err := utils.GCM_decrypt(key, token_encrypted[1], token_encrypted[0], nil)
		cspan.Finish()
		if decrypt_err != nil {
			c.AbortWithStatusJSON(401, gin.H{"reason": "Unauthorized"})
			return
		}
		var placeholder map[string]interface{}
		if json_err := json.Unmarshal([]byte(token_data), &placeholder); json_err != nil {
			span.Finish()
			panic(json_err.Error())
		}
		user_data := &models.User{}

		cspan = tracer.StartSpan("cache find user by uid",
			opentracing.ChildOf(span.Context()),
		)
		entry, cache_err := cache.Get(placeholder["uid"].(string))
		cspan.Finish()
		if cache_err == nil {
			span.Finish()
			c.String(200, entry)
			return
		}

		cspan = tracer.StartSpan("db find user by uid",
			opentracing.ChildOf(span.Context()),
		)
		find_err := userdb.FindByUid("_id", placeholder["uid"].(string), user_data)
		cspan.Finish()
		if find_err != nil {
			span.Finish()
			c.AbortWithStatusJSON(404, gin.H{"reason": "User not found"})
			return
		}
		user_detail := "{\"uid\": \"" + user_data.Uid.Hex() +
			"\", \"username\": \"" + user_data.Username +
			"\", \"email\": \"" + user_data.Email +
			"\", \"role\": \"" + user_data.Role +
			"\", \"avatarurl\": \"" + user_data.Avatarurl +
			"\", \"active\": \"" + strconv.FormatBool(user_data.Active) +
			"\", \"screenname\": \"" + user_data.Screenname +
			"\", \"location\": \"" + user_data.Location +
			"\", \"protected\": \"" + strconv.FormatBool(user_data.Protected) +
			"\", \"description\": \"" + user_data.Description +
			"\", \"verified\": \"" + strconv.FormatBool(user_data.Verified) +
			"\"}"
		cache.Set(placeholder["uid"].(string), user_detail)
		c.String(200, user_detail)
		span.Finish()
	})

	// internal create user
	router.POST("/user", func(c *gin.Context) {
		spanCtx, _ := tracer.Extract(opentracing.HTTPHeaders, opentracing.HTTPHeadersCarrier(c.Request.Header))
		span := tracer.StartSpan("create user", ext.RPCServerOption(spanCtx))
		service_name, _ := c.GetQuery("service")
		if service_name == "" {
			c.AbortWithStatusJSON(401, gin.H{"reason": "Missing service name"})
			return
		}
		user_email := c.PostForm("email")
		user_name := c.PostForm("username")
		password := c.PostForm("password")

		cspan := tracer.StartSpan("generate hash",
			opentracing.ChildOf(span.Context()),
		)
		hashed := utils.GenerateFromPassword(password)
		cspan.Finish()

		new_user := models.User{
			Username:       user_name,
			Email:          user_email,
			Password:       hashed,
			Role:           "standard",
			Lastlogin:      time.Now(),
			Creationtime:   time.Now(),
			Avatarurl:      "http://localhost/profile.png",
			Active:         true,
			Screenname:     user_name,
			Location:       "Earth",
			Protected:      false,
			Description:    "Hi please update your profile",
			Verified:       false,
			Followercount:  0,
			Followingcount: 0,
			Likecount:      0,
			Postcount:      0,
			Updatetime:     time.Now(),
		}

		_, create_err := userdb.Create(new_user)

		if create_err != nil {
			span.Finish()
			panic(create_err.Error())
		}
		c.String(200, "created")
	})

	router.GET(SERVICE_NAME+"/profile/:username", func(c *gin.Context) {
		spanCtx, _ := tracer.Extract(opentracing.HTTPHeaders, opentracing.HTTPHeadersCarrier(c.Request.Header))
		span := tracer.StartSpan("get profile", ext.RPCServerOption(spanCtx))

		name := c.Param("username")
		token, token_err := c.Cookie("token")
		if token_err != nil {
			c.AbortWithStatusJSON(401, gin.H{"reason": "unauthorized"})
			return
		}
		splitted := strings.Split(token, "-")
		_, decrypt_err := utils.GCM_decrypt(key, splitted[1], splitted[0], nil)
		if decrypt_err != nil {
			c.AbortWithStatusJSON(401, gin.H{"reason": "unauthorized"})
			return
		}
		result := &models.User{}
		find_err := userdb.Find("username", name, result)
		if find_err != nil {
			span.Finish()
			c.AbortWithStatusJSON(404, gin.H{"reason": "not found"})
			return
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
			"\", \"follower\": \"" + strconv.Itoa(result.Followercount) +
			"\", \"following\": \"" + strconv.Itoa(result.Followingcount) +
			"\", \"post\": \"" + strconv.Itoa(result.Postcount) +
			"\"}"

		c.String(200, user_detail)
		span.Finish()

	})

	// internal delete user
	router.DELETE("/user/:userid", func(c *gin.Context) {

		spanCtx, _ := tracer.Extract(opentracing.HTTPHeaders, opentracing.HTTPHeadersCarrier(c.Request.Header))
		span := tracer.StartSpan("delete user", ext.RPCServerOption(spanCtx))

		uid := c.Param("userid")
		_, delete_err := userdb.Delete(uid)
		if delete_err != nil {
			span.Finish()
			c.AbortWithStatusJSON(404, gin.H{"reason": "not found"})
			return
		}
		cache.Delete(uid)

		c.String(200, "deleted")
		span.Finish()

	})

	return router

}

func main() {

	mongo_layer := helper.NewMongoDatabase()
	userdb := models.NewUserDatabase(mongo_layer)
	redis_service := services.NewRedisService()

	router := setupRouter(userdb, redis_service)
	router.Run(":8080") // listen and serve on 0.0.0.0:8080

}
