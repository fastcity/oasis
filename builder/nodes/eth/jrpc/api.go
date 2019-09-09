package jrpc

import (
	"strconv"

	jsoniter "github.com/json-iterator/go"
	"github.com/ybbus/jsonrpc"
)

var json = jsoniter.ConfigCompatibleWithStandardLibrary

type api struct {
	baseURL  string
	userName string
	pwd      string
}
type ChainApi interface {
	GetBlockHeight() (int64, error)
	GetBlockInfo(int64, interface{}) error
	GetTransactionReceipt(string, interface{}) error
	CreateTransactionData(interface{}, interface{}) (interface{}, error)
	SubmitTransactionData(interface{}) (interface{}, error)
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

func NewChainAPi(url, userName, pwd string) ChainApi {

	return &api{
		baseURL:  url,
		userName: userName,
		pwd:      pwd,
	}
}

func (u *api) getRpcClient() jsonrpc.RPCClient {
	// rpcClient := jsonrpc.NewClientWithOpts(u.baseURL, &jsonrpc.RPCClientOpts{
	// 	CustomHeaders: map[string]string{
	// 		"Authorization": "Basic " + base64.StdEncoding.EncodeToString([]byte(u.userName+":"+u.pwd)),
	// 	},
	// })
	rpcClient := jsonrpc.NewClient(u.baseURL)
	return rpcClient
}

// GetBlockHeight 获取块高
func (u *api) GetBlockHeight() (height int64, err error) {
	var reply string

	err = u.getRpcClient().CallFor(&reply, "eth_blockNumber")
	if err != nil {
		return -1, err
	}
	//0x4d1a35

	// reply = reply[2:]
	h, err := strconv.ParseInt(reply, 0, 32) ////0x4d1a35 写 0 后 他自己判断去除前面的0x

	return h, err
}

func (u *api) GetBlockInfo(height int64, info interface{}) error {

	h := strconv.FormatInt(height, 16)
	h = "0x" + h

	err := u.getRpcClient().CallFor(&info, "eth_getBlockByNumber", h, true)
	if err != nil {
		return err
	}

	return nil
}

// 根据hash 查询交易
func (u *api) GetTransactionReceipt(hash string, txData interface{}) error {

	// var hashdata interface{}

	err := u.getRpcClient().CallFor(&txData, "eth_getTransactionReceipt", hash)
	if err != nil {
		return err
	}

	return nil
}

// createTransactionData 创建未签名事务
func (u *api) CreateTransactionData(input, output interface{}) (interface{}, error) {
	var tx interface{}
	err := u.getRpcClient().CallFor(&tx, "createrawtransaction", input, output)
	if err != nil {
		return nil, err
	}
	return tx, nil
}

func (u *api) SubmitTransactionData(sign interface{}) (interface{}, error) {
	var txid interface{}
	err := u.getRpcClient().CallFor(&txid, "sendrawtransaction", sign)
	if err != nil {
		return nil, err
	}
	return txid, nil
}
