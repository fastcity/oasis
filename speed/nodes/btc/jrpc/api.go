package jrpc

import (
	"encoding/base64"

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
	GetBlockInfo(int64) (map[string]interface{}, error)
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
	rpcClient := jsonrpc.NewClientWithOpts(u.baseURL, &jsonrpc.RPCClientOpts{
		CustomHeaders: map[string]string{
			"Authorization": "Basic " + base64.StdEncoding.EncodeToString([]byte(u.userName+":"+u.pwd)),
		},
	})
	return rpcClient
}

// GetBlockHeight 获取块高
func (u *api) GetBlockHeight() (height int64, err error) {
	var reply int64

	err = u.getRpcClient().CallFor(&reply, "getblockcount")
	if err != nil {
		return reply, err
	}

	return reply, nil
}

func (u *api) GetBlockInfo(height int64) (map[string]interface{}, error) {

	var hashdata string
	err := u.getRpcClient().CallFor(&hashdata, "getblockhash", height)
	if err != nil {
		return nil, err
	}
	var txData map[string]interface{}
	err = u.getRpcClient().CallFor(&txData, "getblock", hashdata, 2)
	if err != nil {
		return nil, err
	}

	txData["blockHash"] = hashdata

	// b, err := json.Marshal(txData)

	return txData, nil
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
