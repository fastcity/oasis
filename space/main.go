package main

import (
	"century/oasis/space/api"
	"crypto/md5"
	"encoding/hex"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sort"
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

const (
	defaultMaxMemory = 32 << 20 // 32 MB
)

func crypto(data map[string][]string) map[string]string {
	//
	form := map[string]string{}
	for i, v := range data {
		// TODO: 数组
		if len(v) == 1 && v[0] != "" {
			form[i] = v[0]
		}
	}
	return form
}
func sortData(data map[string]string) []string {
	keys := []string{}
	for k, _ := range data {
		keys = append(keys, k)
	}

	sort.Strings(keys)
	return keys

}

func sign(ctx *gin.Context) {
	// form := map[string]string{}

	ctx.Request.ParseMultipartForm(defaultMaxMemory)
	body := ctx.Request.Form

	form := crypto(body)
	sortKey := sortData(form)

	data := ""
	for _, v := range sortKey {
		data += v + "=" + form[v] + "&"
	}
	data = strings.TrimRight(data, "&")

	h := md5.New()
	h.Write([]byte(data))
	cipherStr := h.Sum(nil)

	digest := hex.EncodeToString(cipherStr)

	ctx.Set("signature", digest)
	fmt.Printf("%s\n", digest) // 输出加密结果
	ctx.Next()
}

func createTransferTxData(c *gin.Context) {
	app.SetGinCtx(c).RedirectPsot()
}

func balance(c *gin.Context) {
	fmt.Println("balance")
	app.SetGinCtx(c).RedirectGet()
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
		index.POST("/createTransferTxData", createTransferTxData)
		// index.Any("/balances", any)
		index.Any("", any)
	}

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
func initLog() {
	// 创建记录日志的文件
	f, _ := os.Create("space.log")
	// gin.DefaultWriter = io.MultiWriter(f)

	// 如果需要将日志同时写入文件和控制台，请使用以下代码
	gin.DefaultWriter = io.MultiWriter(f, os.Stdout)
}

func main() {
	flag.StringVar(&env, "env", "dev", "env")
	flag.Parse()

	initConf()
	initLog()

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
