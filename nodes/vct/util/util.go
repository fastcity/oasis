package util

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"math/big"
	"net/http"
	"reflect"
	"strconv"
	"strings"
)

type Util struct {
	BaseURL string
}

type Response struct {
	Result map[string]interface{} `json:"result"`
	Error  ErrInfo                `json:"error"`
}

type ErrInfo struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Data    string `json:"data"`
}

func (u *Util) GetBlockHeight() (height string, err error) {
	url := u.BaseURL + "/chain"

	// body := fmt.Sprintf("accountID=%s&to=%s&amount=3&nonce=%d", from, to, getNonce())
	resp, err := u.apiGet(url)
	if err != nil {
		return "", err
	}
	if resp.Error.Message != "" {
		return "", errors.New(resp.Error.Message)
	}
	fmt.Println("GetBlockHeight", resp)

	h := resp.Result["Height"]

	fmt.Printf(`resp.Result["Height"] %v `, h)
	fmt.Println(`resp.Result["Height"] reflect.TypeOf(h) `, reflect.TypeOf(h))
	reflect.TypeOf(h)
	b, ok := h.(float64)
	if ok {
		fmt.Println(b)
		return strconv.FormatFloat(b, 'f', -1, 64), nil
	}

	return "1", nil
}

func (u *Util) GetBlockInfo(height string) (*Response, error) {
	url := u.BaseURL + "/chain/blocks/" + height

	// body := fmt.Sprintf("accountID=%s&to=%s&amount=3&nonce=%d", from, to, getNonce())
	return u.apiGet(url)
}

// createTransactionData 创建未签名事务
func (u *Util) CreateTransactionData(from, to, tokenKey string, amount *big.Int) (*Response, error) {
	url := u.BaseURL
	if !isEmpty(tokenKey) {
		url = url + "data/token." + tokenKey + "/fund"
	}
	url = url + `data/fund`
	// body := fmt.Sprintf("accountID=%s&to=%s&amount=3&nonce=%d", from, to, getNonce())
	body := fmt.Sprintf("from=%s&to=%s&amount=%s", from, to, amount)
	resp, err := u.apiPost(url, body)
	if err != nil || resp.Error.Message != "" {
		return resp, err
	}
	return resp, err
}

func (u *Util) SubmitTransactionData(rawtx, signStr string) (*Response, error) {
	url := u.BaseURL + `/rawtransaction`
	// body := fmt.Sprintf("accountID=%s&to=%s&amount=3&nonce=%d", from, to, getNonce())
	body := fmt.Sprintf("tx=%s&sig=%s", rawtx, signStr)
	resp, err := u.apiPost(url, body)
	if err != nil || resp.Error.Message != "" {
		return resp, err
	}
	return resp, err
}

func isEmpty(str string) bool {
	if str == "" {
		return true
	}
	return false
}

func (u *Util) apiGet(url string) (*Response, error) {
	res := &Response{}
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	err = json.Unmarshal(body, res)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (u *Util) apiPost(url, requestBody string) (*Response, error) {
	res := &Response{}

	resp, err := http.Post(url, "application/x-www-form-urlencoded", strings.NewReader(requestBody))

	if err != nil {
		return res, err
	}

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	err = json.Unmarshal(body, res)
	if err != nil {
		return res, err
	}

	return res, nil
}
