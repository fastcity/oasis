package gchain

import (
	"errors"
	"fmt"
	"io/ioutil"
	"math/big"
	"net/http"
	"strings"

	"github.com/json-iterator/go"
)

var json = jsoniter.ConfigCompatibleWithStandardLibrary

type api struct {
	BaseURL string
	bytes   []byte
}
type ChainApi interface {
	GetBlockHeight() (height int64, err error)
	GetBlockInfo(height int64) ([]byte, error)
	GetBalance(address string) (balance string, err error)
	CreateTransactionData(from, to, tokenKey string, amount *big.Int) ([]byte, error)
	SubmitTransactionData(rawtx, signStr string) ([]byte, error)
	ToStruct(interface{}) error
	ToResponse() (*Response, error)
}

type Response struct {
	Result map[string]interface{} `json:"result"`
	Error  ErrInfo                `json:"error"`
}

type ErrInfo struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    string `json:"data"`
}

func NewChainAPi(url string) ChainApi {
	return &api{
		BaseURL: url + "/api/v1",
	}
}

// GetBlockHeight 获取块高
func (u *api) GetBlockHeight() (height int64, err error) {
	url := u.BaseURL + "/chain"

	// body := fmt.Sprintf("accountID=%s&to=%s&amount=3&nonce=%d", from, to, getNonce())
	body, err := u.apiGet(url)
	if err != nil {
		return -1, err
	}
	type response struct {
		Result map[string]int64 `json:"result"`
		Error  ErrInfo          `json:"error"`
	}
	resp := &response{}
	err = json.Unmarshal(body, resp)
	if err != nil {
		return -1, err
	}
	// return res, nil
	if resp.Error.Message != "" {
		return -1, errors.New(resp.Error.Message)
	}
	// fmt.Println("GetBlockHeight", resp)

	return resp.Result["Height"], nil

	// fmt.Println(`resp.Result["Height"] reflect.TypeOf(h) `, reflect.TypeOf(h))
	// b, ok := h.(float64)
	// if ok {
	// 	fmt.Println(b)
	// 	return strconv.FormatFloat(b, 'f', -1, 64), nil
	// }

	// return "-1", nil

	// return resp.Result["Height"], nil
}

// GetBlockHeight 获取块高
func (u *api) GetBalance(address string) (balance string, err error) {
	url := u.BaseURL + "/addess/" + address

	// body := fmt.Sprintf("accountID=%s&to=%s&amount=3&nonce=%d", from, to, getNonce())
	body, err := u.apiGet(url)
	if err != nil {
		return "0", err
	}
	type response struct {
		Result map[string]string `json:"result"`
		Error  ErrInfo           `json:"error"`
	}
	resp := &response{}
	err = json.Unmarshal(body, resp)
	if err != nil {
		return "0", err
	}
	// return res, nil
	if resp.Error.Message != "" {
		return "0", errors.New(resp.Error.Message)
	}
	// fmt.Println("GetBlockHeight", resp)

	return resp.Result["Balance"], nil
}

func (u *api) GetBlockInfo(height int64) ([]byte, error) {
	url := fmt.Sprintf("%s/chain/blocks/%d", u.BaseURL, height)

	// body := fmt.Sprintf("accountID=%s&to=%s&amount=3&nonce=%d", from, to, getNonce())
	// return u.apiGet(url)
	// resp := &Response{}
	return u.apiGet(url)
	// err = json.Unmarshal(body, resp)
	// return resp, err
}

// createTransactionData 创建未签名事务
func (u *api) CreateTransactionData(from, to, tokenKey string, amount *big.Int) ([]byte, error) {
	url := u.BaseURL
	if !isEmpty(tokenKey) {
		url = url + "/data/token." + tokenKey + "/fund"
	} else {
		url = url + `/data/fund`
	}

	// body := fmt.Sprintf("accountID=%s&to=%s&amount=3&nonce=%d", from, to, getNonce())
	body := fmt.Sprintf("from=%s&to=%s&amount=%s", from, to, amount)
	// resp, err := u.apiPost(url, body)
	// if err != nil || resp.Error.Message != "" {
	// 	return resp, err
	// }
	// return resp, err
	return u.apiPost(url, body)
}

func (u *api) SubmitTransactionData(rawtx, signStr string) ([]byte, error) {
	url := u.BaseURL + `/rawtransaction`
	// body := fmt.Sprintf("accountID=%s&to=%s&amount=3&nonce=%d", from, to, getNonce())
	body := fmt.Sprintf("tx=%s&sig=%s", rawtx, signStr)
	// resp, err := u.apiPost(url, body)
	// if err != nil || resp.Error.Message != "" {
	// 	return resp, err
	// }
	// return resp, err
	return u.apiPost(url, body)
}

func isEmpty(str string) bool {
	if str == "" {
		return true
	}
	return false
}

func (u *api) apiGet(url string) ([]byte, error) {
	// res := &Response{}
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	return body, err
	// err = json.Unmarshal(body, res)
	// if err != nil {
	// 	return nil, err
	// }
	// return res, nil
}

func (u *api) apiPost(url, requestBody string) ([]byte, error) {
	// res := &Response{}

	resp, err := http.Post(url, "application/x-www-form-urlencoded", strings.NewReader(requestBody))

	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	// err = json.Unmarshal(body, res)
	// if err != nil {
	// 	return res, err
	// }
	u.bytes = body

	return body, err
}

func (u *api) ToStruct(v interface{}) error {
	return json.Unmarshal(u.bytes, v)
}

func (u *api) ToResponse() (*Response, error) {
	res := &Response{}
	err := json.Unmarshal(u.bytes, res)
	if err != nil {
		return res, err
	}
	return res, nil
}
