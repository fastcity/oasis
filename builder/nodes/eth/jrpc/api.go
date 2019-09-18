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
	GetBalance(string, int64) (string, error)

	GetTransactionReceipt(string, interface{}) error
	CreateTransactionData(interface{}, interface{}) (interface{}, error)
	SubmitTransactionData(interface{}) (interface{}, error)

	GetNonce(string) (string, error)
	GetGasPrice() (string, error)
	EstimateGas(params ...interface{}) (string, error)

	// CreateERC20Input(to string, value *big.Int) string
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

	return err
}

func (u *api) GetBalance(address string, height int64) (string, error) {
	h := ""

	if height < 0 {
		h = "latest"
	} else {
		h = strconv.FormatInt(height, 16)
		h = "0x" + h
	}

	var balance string
	err := u.getRpcClient().CallFor(&balance, "eth_getBalance", address, h) // h 块高或者 "latest", "earliest" 或 "pending"

	return balance, err
}

// 根据hash 查询交易
func (u *api) GetTransactionReceipt(hash string, txData interface{}) error {
	err := u.getRpcClient().CallFor(&txData, "eth_getTransactionReceipt", hash)

	return err
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

func (u *api) GetNonce(address string) (string, error) {
	var count string
	err := u.getRpcClient().CallFor(&count, "eth_getTransactionCount", address, "latest")
	if err != nil {
		return "", err
	}

	// nonce, _ := big.NewInt(0).SetString(count, 0) // 根据前缀自己选择

	return count, nil
}

func (u *api) GetGasPrice() (string, error) {
	var price string
	err := u.getRpcClient().CallFor(&price, "eth_gasPrice")
	if err != nil {
		return "", err
	}

	return price, nil
}

// EstimateGas 执行并估算一个交易需要的gas用量
func (u *api) EstimateGas(params ...interface{}) (string, error) {
	var gas string
	err := u.getRpcClient().CallFor(&gas, "eth_estimateGas", params)
	if err != nil {
		return "", err
	}

	return gas, nil
}
