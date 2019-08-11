package api

import (
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

type App struct {
	baseUrl string
	ApiKey  string
	SecKey  string
}
type AppInterface interface {
	SetBaseUrl(base string) AppInterface
	SetApikey(apikey string) AppInterface
	SetSecKey(secKey string) AppInterface

	RedirectGet(path, query string) ([]byte, error)
	RedirectPsot(path, requestBody string) ([]byte, error)
	RedirectAny(ctx *gin.Context) ([]byte, error)
}

func NewApp() AppInterface {
	return &App{}
}

func (a *App) SetBaseUrl(base string) AppInterface {
	a.baseUrl = base
	return a
}

func (a *App) SetApikey(apikey string) AppInterface {
	a.ApiKey = apikey
	return a
}
func (a *App) SetSecKey(secKey string) AppInterface {
	a.SecKey = secKey
	return a
}

//RedirectGet RedirectGet
func (a *App) RedirectGet(path, query string) ([]byte, error) {
	body := path + "?" + query
	resp, err := http.Get(body)
	if err != nil {
		return nil, err
	}

	respbody, err := ioutil.ReadAll(resp.Body)
	return respbody, err
	// return resp
}

func (a *App) RedirectPsot(path, requestBody string) ([]byte, error) {
	body := strings.NewReader(requestBody)
	resp, err := http.Post(path, "application/x-www-form-urlencoded", body)
	if err != nil {
		return nil, err
	}

	respbody, err := ioutil.ReadAll(resp.Body)
	return respbody, err
	// return resp
}

func (a *App) RedirectAny(ctx *gin.Context) ([]byte, error) {
	url := a.baseUrl + ctx.Request.URL.Path
	if ctx.Request.Method == "GET" {
		query := ctx.Request.URL.RawQuery
		return a.RedirectGet(url, query)
	}

	resp, err := http.Post(url, "application/x-www-form-urlencoded", ctx.Request.Body)
	if err != nil {
		return nil, err
	}

	respbody, err := ioutil.ReadAll(resp.Body)
	return respbody, err

	// if err != nil {
	// 	c.JSON(http.StatusOK, gin.H{"code": 40000, "msg": err.Error()})
	// 	return
	// }
	// c.Data(http.StatusOK, "application/json", resp)

}
