package main

import (
	"century/oasis/space/api"
	"flag"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
)

var (
	envprefix  = "space"
	configName = "config"
	env        = "dev"
	db         map[string]string
	app        api.AppInterface
)

func sign(ctx *gin.Context) {
	if ctx.Request.Method == "GET" {
		query := ctx.Request.URL.RawQuery
		andQ := strings.Split(query, "&")

		form := map[string]string{}
		for _, q := range andQ {
			p := strings.Split(q, "=")
			form[p[0]] = p[1]
			// for _, pi := range p {

			// }
		}
	}

}

func balance(c *gin.Context) {
	fmt.Println("balance")

	// baseURL := viper.GetString(env + ".baseUrl")
	// url := baseURL + c.Request.URL.Path
	// query := c.Request.URL.RawQuery
	app.SetGinCtx(c).RedirectGet()
	// app.RedirectGet(url, query, c)
	// resp, err := app.RedirectGet(url, query)
	// if err != nil {
	// 	c.JSON(http.StatusOK, gin.H{"code": 40000, "msg": err.Error()})
	// 	return
	// }

	// // c.Header("Content-Type", "application/json")
	// // c.Writer.Header().Add("Content-Type", "application/json")
	// // c.Writer.Write(resp)
	// c.Data(http.StatusOK, "application/json", resp)

}
func any(c *gin.Context) {
	// fmt.Println("any")
	// app.RedirectAny(c)

	// if err != nil {
	// 	c.JSON(http.StatusOK, gin.H{"code": 40000, "msg": err.Error()})
	// 	return
	// }
	// c.Data(http.StatusOK, "application/json", resp)
	// c.Writer.Write(resp)
}

func setupRouter() *gin.Engine {
	// Disable Console Color
	// gin.DisableConsoleColor()
	r := gin.Default()
	// r.Use()

	// Ping test
	r.GET("/ping", func(c *gin.Context) {
		c.String(http.StatusOK, "pong")
	})

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

	// authorized := r.Group("/", gin.BasicAuth(gin.Accounts{
	// 	"foo":  "bar", // user:foo password:bar
	// 	"manu": "123", // user:manu password:123
	// }))

	// authorized.POST("admin", func(c *gin.Context) {
	// 	user := c.MustGet(gin.AuthUserKey).(string)

	// 	// Parse JSON
	// 	var json struct {
	// 		Value string `json:"value" binding:"required"`
	// 	}

	// 	if c.Bind(&json) == nil {
	// 		db[user] = json.Value
	// 		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	// 	}
	// })

	index := r.Group("/api").Group("/v1")
	{
		index.Use(sign)
		index.GET("/balance", balance)
		index.Any("/balances", any)
		index.Any("", any)
	}

	r.Any("", any)
	// r.GET("/balance", &controllers.BalanceController{}),
	// r.GET("/createTransferTxData", ft, "post:CreateTransferTxData"),
	// r.GET("/submitTx", ft, "post:SubmitTx"),

	// r.GET("/getTxStatus", ft, "get:GetTxStatus"),

	// r.GET("/subscribe", &controllers.AccountController{DB: db}, "post:Subscribe"),

	// r.GET("/account")
	return r
}

func initConf() {

	gopath := os.Getenv("GOPATH")
	path := []string{}

	for _, p := range filepath.SplitList(gopath) {
		// century\oasis\space
		pathConf := filepath.Join(p, "src/century/oasis/space")
		path = append(path, pathConf)
		// viper.AddConfigPath(path)
	}
	path = append(path, ".")
	e := InitViper(envprefix, configName, path)
	if e != nil {
		fmt.Println(e)
	}
}

//InitViper we can set viper which fabric peer is used
func InitViper(envprefix string, filename string, configPath []string) error {
	fmt.Println("envprefix", envprefix, "filename", filename, "configPath", configPath)
	viper.SetEnvPrefix(envprefix)
	viper.AutomaticEnv()
	replacer := strings.NewReplacer(".", "_")
	viper.SetEnvKeyReplacer(replacer)

	for _, c := range configPath {
		viper.AddConfigPath(c)
	}

	viper.SetConfigName(filename) // Name of config file (without extension)
	return viper.ReadInConfig()   // Find and read the config file

}

func main() {
	flag.StringVar(&env, "env", "dev", "env")
	flag.Parse()
	initConf()

	baseURL := viper.GetString(env + ".baseUrl")
	apikey := viper.GetString(env + ".apiKey")
	seckey := viper.GetString(env + ".secKey")

	app = api.NewApp().SetBaseUrl(baseURL).SetApikey(apikey).SetSecKey(seckey)
	r := setupRouter()
	// Listen and Server in 0.0.0.0:8080
	host := viper.GetInt(env + ".host")
	if host == 0 {
		host = 7788
	}
	listen := fmt.Sprintf(":%d", host)

	r.Run(listen)
}
