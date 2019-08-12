package api

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

type App struct {
	baseUrl string
	apiKey  string
	secKey  string
	ctx     *gin.Context

	form map[string]string
}
type AppInterface interface {
	SetBaseUrl(base string) AppInterface
	SetApikey(apikey string) AppInterface
	SetSecKey(secKey string) AppInterface
	SetGinCtx(ctx *gin.Context) AppInterface

	RedirectGet()
	RedirectPsot()
	// RedirectAny()
}

func NewApp() AppInterface {
	return &App{}
}

func (a *App) SetBaseUrl(base string) AppInterface {
	a.baseUrl = base
	return a
}

func (a *App) SetApikey(apikey string) AppInterface {
	a.apiKey = apikey
	return a
}
func (a *App) SetSecKey(secKey string) AppInterface {
	a.secKey = secKey
	return a
}
func (a *App) SetGinCtx(ctx *gin.Context) AppInterface {
	a.ctx = ctx
	return a
}

//RedirectGet RedirectGet
func (a *App) RedirectGet() {

	// path := a.baseUrl + a.ctx.Request.RequestURI
	query := a.ctx.Request.URL.RawQuery

	url := a.baseUrl + a.ctx.Request.RequestURI
	q := a.ctx.Request.Form
	fmt.Println(q)
	signatrue := sign(query, a.secKey)
	body := url + "&signatrue=" + signatrue
	resp, err := http.Get(body)
	if err != nil {
		// c.Status(http.StatusServiceUnavailable)
		a.ctx.JSON(resp.StatusCode, gin.H{"code": 40000, "msg": err.Error()})
		return
	}

	contentLength := resp.ContentLength
	contentType := resp.Header.Get("Content-Type")
	a.ctx.DataFromReader(http.StatusOK, contentLength, contentType, resp.Body, map[string]string{})
	// c.Data(http.StatusOK, contentType, respbody)
}

func (a *App) RedirectPsot() {
	path := a.baseUrl + a.ctx.Request.URL.Path
	requestBody := a.ctx.Request.URL.RawQuery
	body := strings.NewReader(requestBody)
	resp, err := http.Post(path, "application/x-www-form-urlencoded", body) //TODO: 原始type
	if err != nil {
		a.ctx.JSON(resp.StatusCode, gin.H{"code": 40000, "msg": err.Error()})
		return
	}

	// respbody, err := ioutil.ReadAll(resp.Body)

	contentLength := resp.ContentLength
	contentType := resp.Header.Get("Content-Type")
	a.ctx.DataFromReader(http.StatusOK, contentLength, contentType, resp.Body, map[string]string{})
	// return respbody, err
	// return resp
}

func (a *App) RedirectAny() {
	if a.ctx.Request.Method == "GET" {
		a.RedirectGet()
		return
	}

	if a.ctx.Request.Method == "POST" {
		a.RedirectPsot()
		return
	}
	// resp, err := http.Post(url, "application/x-www-form-urlencoded", ctx.Request.Body)
	// if err != nil {
	// 	return nil, err
	// }

	// respbody, err := ioutil.ReadAll(resp.Body)
	// return respbody, err

}

func sign(data, seckey string) string {

	return ""

}
