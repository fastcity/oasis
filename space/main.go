package main

import (
	"century/space/api"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"
	echo "github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
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

func createTransferTxData(c echo.Context) error {
	app.SetGinCtx(c).RedirectPsot()
	return nil
}

func submitTx(c echo.Context) error {
	app.SetGinCtx(c).RedirectPsot()
	return nil
}

func subscribe(c echo.Context) error {
	app.SetGinCtx(c).RedirectPsot()
	return nil
}

//newaccount
func newAccount(c echo.Context) error {
	app.SetGinCtx(c).RedirectPsot()
	return nil
}

func setCallBackUrl(c echo.Context) error {
	app.SetGinCtx(c).RedirectPsot()
	return nil
}

func getAccount(c echo.Context) error {
	fmt.Println("getAccount")
	app.SetGinCtx(c).RedirectGet()
	return nil
}

//getTxStatus

func getTxStatus(c echo.Context) error {
	fmt.Println("getTxStatus")
	app.SetGinCtx(c).RedirectGet()
	return nil
}

func balance(c echo.Context) error {
	fmt.Println("balance")
	app.SetGinCtx(c).RedirectGet()
	return nil
}

func any(e echo.Context) error {
	fmt.Println("any")
	app.SetGinCtx(e).RedirectAny()
	return nil
}

func setupRouter() *echo.Echo {
	// Echo instance
	e := echo.New()

	// Middleware
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(middleware.CORS())

	// Routes
	e.GET("/hello", hello)

	index := e.Group("/api").Group("/v1")
	{
		// index.Use(sign)
		index.GET("/balance", balance)
		index.POST("/createTransferTxData", createTransferTxData)
		index.POST("/submitTx", submitTx)
		index.GET("/getTxStatus", getTxStatus)
		index.GET("/subscribe", subscribe)
		index.GET("/account", getAccount)
		index.POST("/newAccount", newAccount)
		index.PUT("/setCallBackUrl", setCallBackUrl)
		//setCallBackUrl
		// index.Any("/balances", any)

	}
	e.Any("/*", any)
	return e

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
	e := setupRouter()

	// Listen and Server in 0.0.0.0:7688
	host := viper.GetInt(env + ".host")
	if host == 0 {
		host = 7688
	}
	listen := fmt.Sprintf(":%d", host)

	// Start server
	e.Logger.Fatal(e.Start(listen))
	// r.Run(listen)
}

// Handler
func hello(c echo.Context) error {
	return c.String(http.StatusOK, "Hello, World!")
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
