package api

import (
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

type App struct {
	baseUrl string
}
type AppInterface interface {
	RedirectGet(path, query string) ([]byte, error)
	RedirectPsot(path, requestBody string) ([]byte, error)
	RedirectAny(ctx *gin.Context) ([]byte, error)
}

func newApp(base string) AppInterface {
	return &App{
		baseUrl: base,
	}
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
	// return resp
}
