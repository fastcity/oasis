package main

import (
	"century/oasis/space/api"
	"net/http"

	"github.com/gin-gonic/gin"
)

var db map[string]string

func balance(c *gin.Context) {
	// c.String(http.StatusOK, "pong")
	api.RedirectGet(c.Request.URL.Path, c.GetRawData)
}

func setupRouter() *gin.Engine {
	// Disable Console Color
	// gin.DisableConsoleColor()
	r := gin.Default()

	// Ping test
	r.GET("/ping", func(c *gin.Context) {
		c.String(http.StatusOK, "pong")
	})
	// r.Group("/api").Group("/v1").GET("/balance", balance)

	// Get user value
	r.GET("/user/:name", func(c *gin.Context) {
		user := c.Params.ByName("name")
		value, ok := db[user]
		if ok {
			c.JSON(http.StatusOK, gin.H{"user": user, "value": value})
		} else {
			c.JSON(http.StatusOK, gin.H{"user": user, "status": "no value"})
		}
	})

	// Authorized group (uses gin.BasicAuth() middleware)
	// Same than:
	// authorized := r.Group("/")
	// authorized.Use(gin.BasicAuth(gin.Credentials{
	//	  "foo":  "bar",
	//	  "manu": "123",
	//}))
	authorized := r.Group("/", gin.BasicAuth(gin.Accounts{
		"foo":  "bar", // user:foo password:bar
		"manu": "123", // user:manu password:123
	}))

	authorized.POST("admin", func(c *gin.Context) {
		user := c.MustGet(gin.AuthUserKey).(string)

		// Parse JSON
		var json struct {
			Value string `json:"value" binding:"required"`
		}

		if c.Bind(&json) == nil {
			db[user] = json.Value
			c.JSON(http.StatusOK, gin.H{"status": "ok"})
		}
	})

	index := r.Group("/api").Group("/v1")
	{
		index.GET("/balance", balance)
	}
	// r.GET("/balance", &controllers.BalanceController{}),
	// r.GET("/createTransferTxData", ft, "post:CreateTransferTxData"),
	// r.GET("/submitTx", ft, "post:SubmitTx"),

	// r.GET("/getTxStatus", ft, "get:GetTxStatus"),

	// r.GET("/subscribe", &controllers.AccountController{DB: db}, "post:Subscribe"),

	// r.GET("/account")
	return r
}

func main() {
	r := setupRouter()
	// Listen and Server in 0.0.0.0:8080
	r.Run(":7789")
}
