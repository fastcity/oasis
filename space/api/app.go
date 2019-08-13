package api

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"net/http"
	"sort"
	"strings"

	"github.com/gin-gonic/gin"
)

const (
	defaultMaxMemory = 32 << 20 // 32 MB
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
	// query := a.ctx.Request.URL.RawQuery

	url := a.baseUrl + a.ctx.Request.RequestURI
	// q := a.ctx.Request.Form
	// fmt.Println(q)
	signature := sign(a.ctx)
	body := url + "&signature=" + signature
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

	// form := map[string]string{}

	f := a.ctx.Request.Form
	fmt.Println("get params body", f)
	// andQ := strings.Split(body, "&")
	// for _, q := range andQ {
	// 	p := strings.Split(q, "=")
	// 	form[p[0]] = p[1]
	// }

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

func sortKeys(data map[string][]string) []string {
	keys := getKeys(data)
	sort.Strings(keys)
	return keys

}

func getKeys(data map[string][]string) []string {
	keys := []string{}
	for k, _ := range data {
		keys = append(keys, k)
	}
	return keys
}

func sign(ctx *gin.Context) string {
	// form := map[string]string{}

	ctx.Request.ParseMultipartForm(defaultMaxMemory)
	body := ctx.Request.Form

	sortKey := sortKeys(body)

	data := ""
	for _, v := range sortKey {
		// 数组参数
		sort.Strings(body[v])
		for _, fv := range body[v] {
			if fv != "" {
				data += v + "=" + fv + "&"
			}
		}
	}
	data = strings.TrimRight(data, "&")
	fmt.Println("sign data", data)
	h := md5.New()
	h.Write([]byte(data))
	cipherStr := h.Sum(nil)

	digest := hex.EncodeToString(cipherStr)

	// ctx.Set("signature", digest)
	fmt.Printf("%s\n", digest) // 输出加密结果

	return strings.ToLower(digest)
}
