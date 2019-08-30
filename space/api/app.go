package api

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"net/http"
	"sort"
	"strings"

	"github.com/gin-gonic/gin"
	echo "github.com/labstack/echo/v4"
)

const (
	defaultMaxMemory = 32 << 20 // 32 MB
)

type App struct {
	baseUrl string
	apiKey  string
	secKey  string
	ctx     echo.Context

	form map[string]string
}
type AppInterface interface {
	SetBaseUrl(base string) AppInterface
	SetApikey(apikey string) AppInterface
	SetSecKey(secKey string) AppInterface
	SetGinCtx(ctx echo.Context) AppInterface

	RedirectGet()
	RedirectPsot()
	RedirectAny()
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
func (a *App) SetGinCtx(ctx echo.Context) AppInterface {
	a.ctx = ctx
	return a
}

//RedirectGet RedirectGet
func (a *App) RedirectGet() {

	url := a.baseUrl + a.ctx.Request().RequestURI

	_, signature := a.sign()
	body := url + "&signature=" + signature + "&apiKey=" + a.apiKey
	resp, err := http.Get(body)
	if err != nil {
		// c.Status(http.StatusServiceUnavailable)
		a.ctx.JSON(http.StatusServiceUnavailable, gin.H{"code": 40000, "msg": err.Error()})
		return
	}

	// contentLength := resp.ContentLength
	contentType := resp.Header.Get("Content-Type")

	a.ctx.Stream(http.StatusOK, contentType, resp.Body)
}

func (a *App) RedirectPsot() {

	f := a.ctx.Request().Form

	a.ctx.Logger().Debug("get params body", f)
	data, signature := a.sign()
	path := a.baseUrl + a.ctx.Path()
	// requestBody := a.ctx.FormParams()
	requestBody := data + "&signature=" + signature
	body := strings.NewReader(requestBody)
	resp, err := http.Post(path, "application/x-www-form-urlencoded", body) //TODO: 原始type
	if err != nil {
		a.ctx.JSON(http.StatusServiceUnavailable, gin.H{"code": 40000, "msg": err.Error()})
		return
	}

	// respbody, err := ioutil.ReadAll(resp.Body)

	// contentLength := resp.ContentLength
	contentType := resp.Header.Get("Content-Type")
	// a.ctx.DataFromReader(http.StatusOK, contentLength, contentType, resp.Body, map[string]string{})
	a.ctx.Stream(http.StatusOK, contentType, resp.Body)
}

func (a *App) RedirectAny() {
	if a.ctx.Request().Method == "GET" {
		a.RedirectGet()
		return
	}

	if a.ctx.Request().Method == "POST" {
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
	for k := range data {
		keys = append(keys, k)
	}
	return keys
}

func (a *App) sign() (string, string) {
	// form := map[string]string{}

	a.ctx.Request().ParseMultipartForm(defaultMaxMemory)
	body := a.ctx.Request().Form

	body["apiKey"] = []string{a.apiKey}
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
	fmt.Println("========data", data)
	a.ctx.Logger().Debug("-------sign data----------", data)
	h := md5.New()
	h.Write([]byte(data))
	cipherStr := h.Sum(nil)

	digest := hex.EncodeToString(cipherStr)

	// ctx.Set("signature", digest)
	fmt.Println("signature", digest) // 输出加密结果

	return data, strings.ToLower(digest)
}
