package util

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"math/big"
	"net/http"
	"strings"
)

type Util struct {
	BaseURL string
}

type Response struct {
	Result map[string]string `json:"result"`
	Error  ErrInfo           `json:"error"`
}

type ErrInfo struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    string `json:"data"`
}

func (u *Util) GetBlockHeight() (height string, err error) {
	fmt.Println("GetBlockHeight -----------")
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

	return resp.Result["Height"], nil
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
	fmt.Println(string(body))
	err = json.Unmarshal(body, res)
	if err != nil {
		return nil, err
	}
	fmt.Println("res", res)
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

func jsonString(body []byte,res interface) (interface, error) {
resp,	err := json.Unmarshal(body, res)
	if err != nil {
		return resp, err
	}
	return resp, nil
}
