package main

import (
	"century/oasis/builder/nodes/eth/jrpc"
	"century/oasis/builder/nodes/eth/models"
	"century/oasis/builder/nodes/eth/web3"
	"century/oasis/builder/nodes/util"
	"encoding/hex"
	"io/ioutil"
	"math/big"

	"context"
	"flag"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	jsoniter "github.com/json-iterator/go"
	"github.com/shopspring/decimal"
	"github.com/spf13/viper"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.uber.org/zap"
)

var (
	chain              string
	env                string
	json               = jsoniter.ConfigCompatibleWithStandardLibrary
	db                 util.MongoI
	currentBlockNumber = 0
	chainConf          jrpc.ChainApi
	kafka              util.KaInterface
	commondb           = "dynasty"
	chaindb            = "eth"
	chainSymbol        = "ETH"
	confirmedNumber    int64
	logger             *zap.SugaredLogger
	wei                int64 = 1000000000000000000
	gwei               int64 = 100000000
)

// Result 返回结果
type Result struct {
	Code int         `json:"code"`
	Data interface{} `json:"data"`
	Msg  string      `json:"msg"`
}

func main() {

	flag.StringVar(&chain, "chain", "ETH", "chain")
	flag.StringVar(&env, "env", "dev", "env")
	flag.Parse()

	logger = util.NewLogger()
	// viper.SetConfigFile("")
	beforeStart()

	go kafkaListing()

	defer kafka.Close()
	defer db.Close()

	initRouter()
}

func beforeStart() {

	initConf()
	initKafka()
	initDB()

}

func initConf() {

	defaultConf := filepath.Join("../../config/", strings.ToLower(env), "nodes")
	gopath := os.Getenv("GOPATH")
	pathConf := []string{defaultConf, "."}
	for _, p := range filepath.SplitList(gopath) {
		path := filepath.Join(p, "src/century/oasis/builder/config", strings.ToLower(env), "nodes")
		pathConf = append(pathConf, path)
	}

	InitViper(strings.ToLower(chain), strings.ToLower(chain), pathConf)

	confirmedNumber = viper.GetInt64("chain.confirmedNumber")
}

func initDB() {
	db = util.NewDBs(viper.GetString("db.addr"))
	err := db.GetConn()
	if err != nil {
		logger.Error("connect mongo error", err)
		panic(err)
	}

	commondb = viper.GetString("chain.commondb")
	chaindb = viper.GetString("chain.chaindb")

	chainSymbol = strings.ToUpper(chaindb)
}

func initKafka() {
	kafka = util.NewConsumer(viper.GetStringSlice("kafka.service"))
	kafka.SetTopics(viper.GetStringSlice("kafka.topics"))
	kafka.SetKeys(viper.GetStringSlice("kafka.keys"))
}

func initRouter() {

	urlChain := fmt.Sprintf("%s://%s:%s", viper.GetString("node.protocal"), viper.GetString("node.host"), viper.GetString("node.port"))

	username := viper.GetString("node.auth.user")
	pwd := viper.GetString("node.auth.pass")
	chainConf = jrpc.NewChainAPi(urlChain, username, pwd)
	// chainConf.SetFeeUrl(viper.GetString("chain.fee"))
	// // 允许来自所有域名请求
	// r.Header.Add("Access-Control-Allow-Origin", "*")
	// // 设置所允许的HTTP请求方法
	// r.Header.Add("Access-Control-Allow-Methods", "OPTIONS, GET, PUT, POST, DELETE")
	// // 字段是必需的。它也是一个逗号分隔的字符串，表明服务器支持的所有头信息字段.
	// r.Header.Add("Access-Control-Allow-Headers", "x-requested-with, accept, origin, content-type")
	// r.Header.Add("Content-Type", "application/json")

	http.HandleFunc("/api/v1/createTransferTxData", createTransactionDataHandler)
	http.HandleFunc("/api/v1/submitTxData", submitTxDtaHandler)
	http.HandleFunc("/api/v1/getBlockHeight", getBlockHeight)         //getBlockInfo  getBlockHeight
	http.HandleFunc("/api/v1/getBlockInfo", getBlockInfo)             //getBlockInfo  getBlockHeight
	http.HandleFunc("/api/v1/getBlockInfoByHash", getBlockInfoByHash) //getBlockInfo  getBlockHeight
	http.HandleFunc("/api/v1/getFeeRate", getFeeRate)                 //getBlockInfo  getBlockHeight
	http.HandleFunc("/api/v1/getBalance", getBalance)
	http.HandleFunc("/api/v1/history", getHistory)

	url := fmt.Sprintf("%s:%s", viper.GetString("api.host"), viper.GetString("api.port"))

	fmt.Println("listen api:", url, "........")
	err := http.ListenAndServe(url, nil)
	if err != nil {
		fmt.Println("http listen failed.", err)
		panic(err)
	}
}

func JSON(w http.ResponseWriter, data interface{}) error {
	w.Header().Set("Content-Type", "application/json")
	encoder := json.NewEncoder(w)
	return encoder.Encode(data)
}

func escapeString(str, e string) {
	strings.Replace(str, "+", "%2B", -1)
}

func createTransactionDataHandler(w http.ResponseWriter, r *http.Request) {
	// fmt.Println("createTransactionDataHandler------")
	type transfer struct {
		Chain     string
		Coin      string
		From      string `json:"from"`
		To        string
		Value     string
		Nonce     string
		Amount    *big.Int
		NonceBig  *big.Int
		TokenKey  string
		FeeRate   string `json:"feeRate"`
		RequestID string `json:"requestId"`
	}
	res := &Result{}
	defer func(res *Result) {
		w.Header().Add("Content-Type", "application/json")
		ba, err := json.Marshal(res)
		if err != nil {
			res.Code = 40000
			res.Msg = err.Error()
		}
		w.Write(ba)
	}(res)

	if r.Method == http.MethodPost {
		tf := &transfer{}
		switch r.Header.Get("Content-Type") {
		case "application/json":
			body, _ := ioutil.ReadAll(r.Body)
			err := json.Unmarshal(body, tf)
			if err != nil {
				res.Code = 40000
				res.Msg = err.Error()
				return
			}

		case "application/x-www-form-urlencoded":
			tf.From = r.PostFormValue("from")
			tf.To = r.PostFormValue("to")
			tf.TokenKey = r.PostFormValue("tokenKey")
			tf.RequestID = r.PostFormValue("requestId")
			tf.Value = r.PostFormValue("value")
			tf.Nonce = r.PostFormValue("nonce")
			tf.FeeRate = r.PostFormValue("feeRate")
			tf.Coin = r.PostFormValue("coin")

		default:
			w.WriteHeader(406)
			res.Code = 406
			res.Msg = "not support Content-Type"
			return
		}

		value64, err := decimal.NewFromString(tf.Value)

		if err != nil {
			res.Code = 40000
			res.Msg = "Invalid amount" + err.Error()
			return
		}

		value := value64.Mul(decimal.New(wei, 0))

		amount := big.NewInt(value.IntPart())

		tf.Amount = amount

		tf.Chain = chainSymbol
		if tf.Coin == "ERC20" {
			b, _ := json.Marshal(tf)
			res1 := createERC20TransactionData(b)
			res.Code = res1.Code
			res.Data = res1.Data
			res.Msg = res1.Msg
			return
		}
		tf.TokenKey = "-"
		tf.Coin = "ETH"

		nonceChain, _ := chainConf.GetNonce(tf.From)
		nonceChainInt := web3.HexToBigint(nonceChain)

		if tf.Nonce != "" {
			nonce, ok := big.NewInt(0).SetString(tf.Nonce, 0)

			if !ok || (amount.IsUint64() && amount.Uint64() == 0) {
				res.Code = 40000
				res.Msg = "Invalid nonce"
				return
			}
			tf.NonceBig = nonce

			if nonceChainInt.Cmp(nonce) < 0 {
				res.Code = 40000
				res.Msg = "Invalid nonce"
				return
			}
		} else {
			tf.NonceBig = nonceChainInt
		}
		var gasPrice *big.Int

		if tf.FeeRate == "" {
			gasPriceA, _ := chainConf.GetGasPrice() // 单位 wei
			gasPrice = web3.HexToBigint(gasPriceA)
			tf.FeeRate = gasPrice.String()
		} else {
			gasPriceD, err := decimal.NewFromString(tf.FeeRate)

			if err != nil {
				res.Code = 40000
				res.Msg = "Invalid FeeRate" + err.Error()
				return
			}

			gasPrice = big.NewInt(gasPriceD.IntPart())
			if gasPrice.Cmp(big.NewInt(200)) < 1 { // 用户传的可能是wei gwei 小于200 认为是gwei 否则认为为wei
				gasPrice = web3.Gweitowei(gasPrice)
			} else {
				gasPrice, _ = big.NewInt(0).SetString(tf.FeeRate, 0)
			}

		}

		gasStruct := map[string]string{
			"from":  tf.From,
			"to":    tf.To,
			"value": web3.BigToHex(tf.Amount, true),
		}
		gasHex, _ := chainConf.EstimateGas(gasStruct) // 单位 wei hex

		gas := web3.HexToBigint(gasHex)

		fee := web3.WeitoEther(gas.Div(gas, gasPrice))

		_, err = db.GetCollection(chaindb, "transfertochains").InsertOne(context.Background(),
			bson.M{"chain": chainSymbol, "coin": tf.Coin, "from": tf.From, "to": tf.To, "tokenKey": tf.TokenKey,
				"value": tf.Value, "fee": fee, "requestId": tf.RequestID})

		if err != nil {
			logger.Error("InsertOne transfertochains error", err)
		}

		txData := map[string]interface{}{
			"to":       tf.To,
			"value":    web3.BigToHex(tf.Amount, true),
			"nonce":    web3.BigToHex(tf.NonceBig, true),
			"gasPrice": web3.BigToHex(gasPrice, true),
			"gas":      gasHex,
			"data":     "",
		}

		res.Code = 0
		res.Data = map[interface{}]interface{}{
			"txData": txData,
		}
		return
	}

	w.WriteHeader(405)
	res.Code = 405
	res.Msg = "method not allow"

	return
}

func createERC20TransactionData(body []byte) *Result {
	type transfer struct {
		Chain     string
		Coin      string
		From      string `json:"from"`
		To        string
		Value     string
		Nonce     string
		Amount    *big.Int
		NonceBig  *big.Int
		TokenKey  string
		FeeRate   string `json:"feeRate"`
		RequestID string `json:"requestId"`
	}
	res := &Result{}

	tf := &transfer{}
	err := json.Unmarshal(body, tf)
	if err != nil {
		res.Code = 40000
		res.Msg = err.Error()
		return res
	}

	tf.From = strings.ToLower(tf.From)
	tf.To = strings.ToLower(tf.To)
	tf.TokenKey = strings.ToLower(tf.TokenKey)

	value64, err := decimal.NewFromString(tf.Value)
	if err != nil {
		res.Code = 40000
		res.Msg = "Invalid amount" + err.Error()
		return res
	}

	value := value64.Mul(decimal.New(wei, 0))
	amount := big.NewInt(value.IntPart())

	tf.Amount = amount
	tf.Chain = chainSymbol

	nonceChain, _ := chainConf.GetNonce(tf.From)
	nonceChainInt := web3.HexToBigint(nonceChain)

	if tf.Nonce != "" {
		nonce, ok := big.NewInt(0).SetString(tf.Nonce, 0)

		if !ok || (amount.IsUint64() && amount.Uint64() == 0) {
			res.Code = 40000
			res.Msg = "Invalid nonce"
			return res
		}
		tf.NonceBig = nonce

		if nonceChainInt.Cmp(nonce) < 0 {
			res.Code = 40000
			res.Msg = "Invalid nonce"
			return res
		}
	} else {
		tf.NonceBig = nonceChainInt
	}
	var gasPrice *big.Int

	if tf.FeeRate == "" {
		gasPriceA, _ := chainConf.GetGasPrice() // 单位 wei
		gasPrice = web3.HexToBigint(gasPriceA)
		tf.FeeRate = gasPrice.String()
	} else {
		gasPriceD, err := decimal.NewFromString(tf.FeeRate)

		if err != nil {
			res.Code = 40000
			res.Msg = "Invalid FeeRate" + err.Error()
			return res
		}

		gasPrice = big.NewInt(gasPriceD.IntPart())
		if gasPrice.Cmp(big.NewInt(200)) < 1 { // 用户传的可能是wei gwei 小于200 认为是gwei 否则认为为wei
			gasPrice = web3.Gweitowei(gasPrice)
		} else {
			gasPrice, _ = big.NewInt(0).SetString(tf.FeeRate, 0)
		}

	}

	data := web3.CreateERC20Input(tf.To, tf.Amount)

	gasStruct := map[string]string{
		"from":     tf.From,
		"to":       tf.TokenKey,
		"value":    web3.BigToHex(big.NewInt(0), true),
		"nonce":    web3.BigToHex(tf.NonceBig, true),
		"gasPrice": web3.BigToHex(gasPrice, true),
		"data":     data,
	}

	// gasStruct := map[string]string{
	// 	"from":     "0xfef1f3dae07a41b00d785cf5af55c57f9145bca0",
	// 	"to":       "0x73dac9535f97d95790b4626e8688ec7d31402c6a",
	// 	"value":    "0x0",
	// 	"nonce":    "0x44",
	// 	"gasPrice": "0x3b9aca00",
	// 	"data":     "0xa9059cbb0000000000000000000000000860123e5bc9bc6f40789e6f2929f7fdf35643ff000000000000000000000000000000000000000000000000016345785d8a0000",
	// }

	// 	data:0xa9059cbb00000000000000000000000x0860123e5bc9bc6f40789e6f2929f7fdf35643ff000000000000000000000000000000000000000000000000016345785d8a0000
	// from:0xfef1f3dae07a41b00d785cf5af55c57f9145bca0
	// gasPrice:0x3b9aca00
	// nonce:0x42
	// to:0x73dac9535f97d95790b4626e8688ec7d31402c6a
	// value:0x0

	gasHex, err := chainConf.EstimateGas(gasStruct) // 单位 wei hex
	if err != nil {
		logger.Error("---------EstimateGas err", err, gasStruct)
		res.Code = 40000
		res.Msg = "EstimateGas error" + err.Error()
		return res
	}

	gas := web3.HexToBigint(gasHex)

	fee := web3.WeitoEther(gas.Div(gas, gasPrice))

	_, err = db.GetCollection(chaindb, "transfertochains").InsertOne(context.Background(),
		bson.M{"chain": chainSymbol, "coin": tf.Coin, "from": tf.From, "to": tf.To, "tokenKey": tf.TokenKey,
			"value": tf.Value, "fee": fee, "requestId": tf.RequestID})

	if err != nil {
		logger.Error("InsertOne transfertochains error", err)
	}

	txData := map[string]interface{}{
		"to":       tf.To,
		"value":    web3.BigToHex(tf.Amount, true),
		"nonce":    web3.BigToHex(tf.NonceBig, true),
		"gasPrice": web3.BigToHex(gasPrice, true),
		"gas":      gasHex,
		"data":     data,
	}

	res.Code = 0
	res.Data = map[string]interface{}{
		"txData": txData,
	}
	return res

}

func submitTxDtaHandler(w http.ResponseWriter, r *http.Request) {
	requestID := r.PostFormValue("requestId")
	signedRawTx := r.PostFormValue("signedRawTx")

	res := submitTx(map[string]string{
		"requestId":   requestID,
		"signedRawTx": signedRawTx,
	})

	ba, _ := json.Marshal(res)
	w.Header().Add("Content-Type", "application/json")
	w.Write(ba)

}

func getBlockInfo(w http.ResponseWriter, r *http.Request) {

	height := r.FormValue("height")

	h, _ := strconv.ParseInt(height, 10, 0)

	fmt.Println("-------------- getBlockInfo", h)

	var data map[string]interface{}

	err := chainConf.GetBlockInfo(h, &data)
	if err != nil {
		fmt.Println("--------------err", err)
	}

	res := &Result{
		Code: 0,
		Data: map[interface{}]interface{}{
			"data": data,
		},
	}

	b, _ := json.Marshal(res)
	w.Header().Add("Content-Type", "application/json")

	fmt.Fprintln(w, string(b))
}

func getBlockInfoByHash(w http.ResponseWriter, r *http.Request) {
	fmt.Println("-------------- getBlockInfoByHash")

	hash := r.FormValue("hash")

	var data map[string]interface{}

	err := chainConf.GetTransactionReceipt(hash, &data)
	if err != nil {
		fmt.Println("--------------err", err)
	}

	res := &Result{
		Code: 0,
		Data: map[interface{}]interface{}{
			"data": data,
		},
	}

	b, _ := json.Marshal(res)
	w.Header().Add("Content-Type", "application/json")

	fmt.Fprintln(w, string(b))
}

func getFeeRate(w http.ResponseWriter, r *http.Request) {
	fmt.Println("-------------- getFeeRate")

	res := &Result{}
	defer func(res *Result) {
		w.Header().Add("Content-Type", "application/json")
		ba, err := json.Marshal(res)
		if err != nil {
			res.Code = 40000
			res.Msg = err.Error()
		}
		w.Write(ba)
	}(res)

	avg, err := chainConf.GetGasPrice()
	if err != nil {
		res.Code = 40000
		res.Msg = err.Error()
		return
	}
	avgb := web3.HexToBigint(avg)

	avgRate := decimal.NewFromBigInt(avgb, 0).Div(decimal.New(1000000000, 0))

	fastRate := avgRate.Mul(decimal.New(2, 0))
	slowRate := avgRate.Div(decimal.New(2, 0))

	res.Code = 0
	res.Data = map[string]interface{}{
		"data": map[string]interface{}{
			"average": avgRate.IntPart(),
			"fast":    fastRate.IntPart(),
			"slow":    slowRate.IntPart(),
		},
	}

	return
}

func getBlockHeight(w http.ResponseWriter, r *http.Request) {
	fmt.Println("-------------- getBlockHeight")
	h, err := chainConf.GetBlockHeight()
	if err != nil {
		fmt.Println("--------------err", err)
	}

	res := &Result{
		Code: 0,
		Data: map[interface{}]interface{}{
			"Height": h,
		},
	}

	b, _ := json.Marshal(res)
	w.Header().Add("Content-Type", "application/json")

	fmt.Fprintln(w, string(b))
}

func getBalance(w http.ResponseWriter, r *http.Request) {
	res := &Result{}
	defer func(res *Result) {
		w.Header().Add("Content-Type", "application/json")
		ba, _ := json.Marshal(res)
		w.Write(ba)
	}(res)

	// ba, _ := json.Marshal(res)
	address := r.FormValue("address")
	fmt.Println("-------------- getBalance", address)
	if address == "" {
		res.Code = 40000
		res.Msg = "address empty"
		return
	}

	b, err := chainConf.GetBalance(address, -1)
	if err != nil {
		res.Code = 40000
		res.Msg = err.Error()
		return
	}

	bb, _ := strconv.ParseInt(b, 0, 32)

	res.Code = 0
	res.Data = map[string]interface{}{
		"total": string(bb),
	}

	// w.Write(ba)

	// // // 允许来自所有域名请求
	// w.Header().Add("Access-Control-Allow-Origin", "*")
	// // 设置所允许的HTTP请求方法
	// w.Header().Add("Access-Control-Allow-Methods", "OPTIONS, GET, PUT, POST, DELETE")
	// // 字段是必需的。它也是一个逗号分隔的字符串，表明服务器支持的所有头信息字段.
	// w.Header().Add("Access-Control-Allow-Headers", "x-requested-with, accept, origin, content-type")
	// w.Header().Add("Content-Type", "application/json")
	// // fmt.Fprintln(w, string(ba))
	// fmt.Println("ba", ba)
	// w.Write(ba)
}

func getHistory(w http.ResponseWriter, r *http.Request) {
	res := &Result{}
	defer func(res *Result) {
		ba, _ := json.Marshal(res)
		w.Header().Add("Content-Type", "application/json")
		w.Write(ba)
	}(res)
	address := r.FormValue("address")
	if address == "" {
		res.Code = 40000
		res.Msg = "address empty"
		return
	}
	pageIndex := r.PostFormValue("pageIndex")

	pageSize := r.PostFormValue("pageIndex")

	pi := 1
	ps := 100

	if pageIndex != "" {
		pi, _ = strconv.Atoi(pageIndex)
	}
	if pi <= 0 {
		pi = 1
	}

	if pageSize == "" {
		ps, _ = strconv.Atoi(pageSize)
	}
	if ps <= 0 {
		ps = 100
	}

	op := options.Find().SetSkip(int64(ps * (pi - 1))).SetLimit(int64(ps)).SetSort(bson.M{"createdAt": -1})

	where := bson.D{
		{
			"$or", bson.A{
				bson.D{{"from", address}},
				bson.D{{"to", address}},
			},
		},
	}

	total, _ := db.GetCollection(chaindb, "transferfromchains").CountDocuments(context.Background(), where)

	result, err := db.GetCollection(chaindb, "transferfromchains").Find(context.Background(), where, op)

	if err != nil {
		res.Code = 40000
		res.Msg = err.Error()
		return
	}
	// var history map[string]interface{}
	// result.Decode(&history)
	var history []map[string]interface{}

	defer result.Close(context.Background())

	for result.Next(context.Background()) {
		var historyOne map[string]interface{}
		result.Decode(&historyOne)
		history = append(history, historyOne)
	}

	count := len(history)
	// var history []map[string]interface{}
	// result.Decode(&history)
	res.Code = 0
	res.Data = map[string]interface{}{
		"items": history,
		"count": count,
		"total": total,
	}
}

//InitViper we can set viper which fabric peer is used
func InitViper(envprefix string, filename string, configPath []string) error {
	fmt.Println("envprefix", envprefix, "filename", filename, "configPath", configPath)
	viper.SetEnvPrefix(envprefix)
	viper.AutomaticEnv()
	replacer := strings.NewReplacer(".", "_")
	viper.SetEnvKeyReplacer(replacer)

	for _, c := range configPath {
		viper.AddConfigPath(c)
	}

	viper.SetConfigName(filename) // Name of config file (without extension)
	return viper.ReadInConfig()   // Find and read the config file

}

func kafkaListing() {
	go kafka.ReciveMsg()

	for {
		for k, v := range kafka.GetMsg() {
			switch k {
			case "TX":
				txHand(v)
			case "SUBMIT_TRANSFER":
				submitTxKakfka(v)
			}

		}
	}

	// kafka.ReciveMsg()
}

func submitTxKakfka(body []byte) {
	var bd map[string]string

	err := json.Unmarshal(body, &bd)
	if err != nil {
		logger.Error("submitTxKakfka error", err)
		return
	}
	submitTx(bd)
}

func submitTx(body map[string]string) (res *Result) {

	requestID := body["requestId"]
	signedRawTx := body["signedRawTx"]

	if res == nil {
		res = &Result{}
	}
	// res := &Result{}
	defer func(res *Result) {
		if res.Code != 0 {
			setSendTransactionError(requestID, res.Msg)
		}
	}(res)

	txid, err := chainConf.SubmitTransactionData(signedRawTx)
	if err != nil {
		res.Code = 40000
		res.Msg = err.Error()
		return
	}

	res.Code = 0
	res.Data = map[string]interface{}{
		"txid": txid,
	}

	setSendTransactionTxid(requestID, txid)
	return
}

func txHand(msg []byte) {

	tfcs := paserTx(msg)
	for _, tfc := range tfcs {
		responseNewTx(tfc)
	}

	newBlockNotify(string(msg))
}

func paserTx(msg []byte) []models.TransferFromChain {
	data := string(msg)

	// logTransferSig := []byte("Transfer(address,address,uint256)")
	// LogApprovalSig := []byte("Approval(address,address,uint256)")
	// logTransferSigHash := crypto.Keccak256Hash(logTransferSig)// transfer   0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef
	// logApprovalSigHash := crypto.Keccak256Hash(LogApprovalSig)// token 发行 0x8c5be1e5ebec7d5bd14f71427d1e84f3dd0314c0f7b2291e5b200ac8c7c3b925

	//合约事件,目前只关注合约转账
	eventTransfer := "0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef"
	// eventApproval := "0x8c5be1e5ebec7d5bd14f71427d1e84f3dd0314c0f7b2291e5b200ac8c7c3b925"

	tfcs := []models.TransferFromChain{}
	h, _ := strconv.ParseInt(data, 10, 0)

	// h := big.NewInt(0).SetBytes(msg).Int64()

	fmt.Println(chaindb, "-+++++++++", data, h)
	// h := string(msg)
	type res struct {
		Result *models.Blocks `json:"result"`
	}
	// fmt.Println("--------------- blockInfo", string(blockInfos))
	curor, err := db.GetCollection(chaindb, "transactions").Find(context.Background(), bson.M{"blockHeight": h})

	if err != nil {
		logger.Error("get transferaction error")
		return tfcs
	}

	defer curor.Close(context.Background())
	for curor.Next(context.Background()) {
		tx := models.Transaction{}
		if err := curor.Decode(&tx); err != nil {
			logger.Error("get transferaction error", err)
		}

		if len(tx.Logs) > 0 {
			for _, log := range tx.Logs {
				// fmt.Println(`log["topics"].(type):`, reflect.TypeOf(log["topics"]), ` log["data"].(type)`, reflect.TypeOf(log["data"])) // primitive.A
				if logtopics, ok := log["topics"]; ok {
					switch topics := logtopics.(type) {
					case primitive.A:
						var data string

						switch datas := log["data"].(type) {
						case string:
							data = datas
						default:
							logger.Error("log[data]  error", log)
						}

						if len(topics) > 1 {
							switch topic := topics[0].(type) {
							case string:
								if strings.ToLower(topic) == eventTransfer { // 只关注转账事件
									if len(topics) == 1 {
										tfc := models.TransferFromChain{
											Chain: "ETH",
											Coin:  "ERC721",
										}

										tfc.From = "0x" + data[26:26+40]
										tfc.To = "0x" + data[26+40+24:26+40+24+40]

										tokenID := data[26+40+24+40+24:]

										tfc.TokenKey = tx.To
										tfc.BlockHeight = tx.BlockHeight
										tfc.Txid = tx.Txid
										tfc.Status = tx.Status
										tfc.Value = strings.TrimRight(tokenID, "0")
										tfc.BlockTime = tx.BlockTime

										value, _ := strconv.ParseInt(strings.TrimRight(tokenID, "0"), 0, 32)

										tfc.Value = string(value)

										tfcs = append(tfcs, tfc)
									} else {
										// "address" : "0xFF603F43946A3A28DF5E6A73172555D8C8b02386", tokenKey ---> contract address
										// "topics" : [
										// 	"0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef",
										// 	"0x0000000000000000000000006cc5f688a315f3dc28a7781717a9a798a59fda7b",  // from
										// 	"0x000000000000000000000000932a764b48542d8d362f2a05e43388e0f548f7c8"   // to
										// ],
										// "data" : "0x000000000000000000000000000000000000000000000531d1d57a8605af8000", // value

										tfc := models.TransferFromChain{
											Chain: "ETH",
											Coin:  "ERC20",
										}

										tfc.From = web3.HexToAddr(topics[1])

										tfc.To = web3.HexToAddr(topics[2])

										tfc.TokenKey = tx.To
										tfc.BlockHeight = tx.BlockHeight
										tfc.Txid = tx.Txid
										tfc.Status = tx.Status
										tfc.Value = strings.TrimRight(data[2:], "0")
										tfc.BlockTime = tx.BlockTime

										by, _ := hex.DecodeString(strings.TrimRight(data[2:], "0"))
										value := big.NewInt(0).SetBytes(by)
										tfc.Value = value.String()

										tfcs = append(tfcs, tfc)
									}

								}
							default:
								logger.Error("contract get error topics [0]", topics[0])
							}

						}
					}
				} else {
					logger.Error("log not find topics", log)
				}
			}
		}
		if tx.Value > 0 {
			// 普通转账
			tfc := models.TransferFromChain{
				Chain: "ETH",
				Coin:  "ETH",

				BlockHeight: tx.BlockHeight,
				BlockTime:   tx.BlockTime,
				Txid:        tx.Txid,
				From:        tx.From,
				To:          tx.To,
				TokenKey:    "-",
				Value:       string(tx.Value),
				Status:      tx.Status,
			}
			tfcs = append(tfcs, tfc)
		}

	}
	return tfcs
}

func responseNewTx(tfc models.TransferFromChain) {

	newTranferFromChain(tfc)

}

func setSendTransactionTxid(requestID, txid string) {
	id, _ := primitive.ObjectIDFromHex(requestID)
	where := bson.M{"_id": id}
	ctx := context.Background()
	result := db.GetCollection(commondb, "transfers").FindOne(ctx, where)
	if result == nil {
		logger.Error("transfer not found!")
		return
	}
	rawBytes, _ := result.DecodeBytes()
	account := rawBytes.Lookup("_account").String() //TODO: 有objectId ?

	// 更新转账操作记录
	updateStr := bson.M{"$set": bson.M{"txid": txid, "code": 16, "status": `TXID`, "updatedAt": time.Now().Unix()}, "$push": bson.M{"logs": `TX_HASH at: ${Date.now()}`}}
	db.GetCollection(commondb, "transfers").FindOneAndUpdate(ctx, where, updateStr)
	// // 更新转账账单记录
	updateStr1 := bson.M{"$set": bson.M{"code": 16, "txid": txid, "updatedAt": time.Now().Unix()}}
	tc := db.GetCollection(chaindb, "transferstochains").FindOneAndUpdate(ctx, bson.M{"requestId": requestID}, updateStr1)

	if tc.Err() != nil {
		logger.Error("error", tc.Err())
	}
	tcdecode := models.TransferToChain{}
	tc.Decode(&tcdecode)

	notifyData := map[string]string{
		"status":    "SUBMIT_TRANSACTION_TO_CHAIN",
		"requestId": requestID,
		"tfcId":     tcdecode.ID.Hex(),
		"txid":      txid,
	}
	sendNotify("TRANSFER_ACTION", account, tcdecode.From, notifyData) //TODO : _account
}

func setSendTransactionError(requestID, msg string) {

	id, _ := primitive.ObjectIDFromHex(requestID)
	where := bson.M{"_id": id}
	ctx := context.Background()

	exist := db.GetCollection(chaindb, "transfers").FindOne(ctx, where)
	if exist == nil {
		logger.Error("transfer not found!")
		return
	}

	// TODO: 时间类型
	// 更新转账操作记录
	updateStr := bson.M{"$set": bson.M{"code": -1, "status": `ERROR`, "updatedAt": time.Now().Unix()}, "$push": bson.M{"logs": `TX_ERROR at: ` + time.Now().String() + msg}}
	transfer := db.GetCollection(commondb, "transfers").FindOneAndUpdate(ctx, where, updateStr)

	// 更新转账账单记录
	tc := db.GetCollection(chaindb, "transferstochains").FindOneAndDelete(ctx, bson.M{"requestId": requestID})

	if tc.Err() != nil {
		logger.Error("error", tc.Err())
	}

	tcdecode := models.TransferToChain{}
	tc.Decode(&tcdecode)
	// // 构造消息
	rawTf, _ := transfer.DecodeBytes()
	account := rawTf.Lookup("_account").String() //TODO： 验证

	// await this.sendNotify('TRANSFER_ACTION', notifyData, _account)
	notifyData := map[string]string{
		"status":    "SUBMIT_TRANSACTION_ERROR",
		"requestId": requestID,
		"msg":       msg,
	}

	sendNotify("TRANSFER_ACTION", account, tcdecode.From, notifyData) //TODO: _account
}

func newTranferFromChain(tfc models.TransferFromChain) {
	tx := tfc
	now := primitive.DateTime(time.Now().Unix() * 1000)
	nowString := time.Now().String()

	tx.CreatedAt = now
	tx.UpdatedAt = nowString
	op := options.FindOneAndUpdate().SetUpsert(true)
	ctx := context.Background()
	where := bson.M{"txid": tx.Txid}
	ttcResult := db.GetCollection(chaindb, "transferTochains").FindOne(ctx, where)

	if ttcResult.Err() != nil && ttcResult.Err() != mongo.ErrNoDocuments {
		logger.Error(">>>>>>>>>>>>>ttcResult", ttcResult.Err())
		return
	}

	var updateStr bson.M
	if ttcResult != nil && ttcResult.Err() != nil {
		ttc := models.TransferToChain{}
		ttcResult.Decode(&ttc)
		tx.ID = ttc.ID
		tx.RequestId = ttc.RequestId

		updateStr1 := bson.M{
			"$set":  bson.M{"code": 32, "status": "FROM_CHAIN", "updatedAt": nowString},
			"$push": bson.M{"logs": `FROM_CHAIN at: ` + nowString},
		}

		id, _ := primitive.ObjectIDFromHex(ttc.RequestId)
		db.GetCollection(commondb, "transfers").FindOneAndUpdate(ctx, bson.M{"_id": id}, updateStr1)

		// 添加地址到订阅

		addSubscribesHandle(ttc.From, ttc.ID.Hex())
	}

	if tx.ID.IsZero() {
		tx.ID = primitive.NewObjectID()
	}

	updateStr = bson.M{"$set": tx}

	if confirmedNumber == 0 {
		db.GetCollection(chaindb, "transferfromchains").FindOneAndUpdate(context.Background(), bson.M{"txid": tx.Txid}, updateStr, op)
	} else {
		db.GetCollection(chaindb, "transferconfirmings").FindOneAndUpdate(context.Background(), bson.M{"txid": tx.Txid}, updateStr, op)
	}

	onchain(tx.From, "OUT", tx)

	onchain(tx.To, "IN", tx)
}

//
func newBlockNotify(blockNumber string) {
	ctx := context.Background()
	number, _ := strconv.ParseInt(blockNumber, 10, 0)
	// number := s
	safeBlockNumber := number - confirmedNumber + 1

	op := options.Find().SetLimit(1024)
	result, err := db.GetCollection(chaindb, "transferconfirmings").Find(ctx, bson.M{}, op)
	if err != nil {
		logger.Error("transfer not found!")
		return
	}
	if result.Next(ctx) {
		ttc := models.TransferConfirming{}
		result.Decode(&ttc)

		confirmdata := map[string]interface{}{
			"id":        ttc.ID.Hex(),
			"requestId": ttc.RequestId,
			"txid":      ttc.Txid,
		}

		isfinish := ttc.BlockHeight <= safeBlockNumber
		confirmedNum := number - ttc.BlockHeight + 1
		if isfinish {
			// now := primitive.DateTime(time.Now().Unix())
			op := options.FindOneAndUpdate().SetUpsert(true)
			db.GetCollection(chaindb, "transferfromchains").FindOneAndUpdate(context.Background(), bson.M{"_id": ttc.ID}, bson.M{"$set": ttc}, op)

			// if tfc.Err() == mongo.ErrNoDocuments {
			// 	db.GetCollection(chaindb, "transferfromchains").FindOneAndUpdate(context.Background(), bson.M{"_id": ttc.ID}, bson.M{"$set": bson.M{"createdAt": now}})
			// }
			if ttc.RequestId != "" {
				id, _ := primitive.ObjectIDFromHex(ttc.RequestId)
				db.GetCollection(commondb, "transfers").FindOneAndDelete(context.Background(), bson.M{"_id": id})
			}

			switch froms := ttc.From.(type) {
			case []string:
				for _, from := range froms {
					finish(from, "OUT", confirmdata)

				}
			case string:
				finish(froms, "OUT", confirmdata)
			}

			switch tos := ttc.To.(type) {
			case []string:
				for _, to := range tos {
					finish(to, "IN", confirmdata)
				}
			case string:
				finish(tos, "IN", confirmdata)
			}

		} else {

			switch froms := ttc.From.(type) {

			case []string:
				for _, from := range froms {
					confirm(from, "OUT", confirmedNum, confirmdata)

				}
			case string:
				confirm(froms, "OUT", confirmedNum, confirmdata)
			}

			switch tos := ttc.To.(type) {
			case []string:
				for _, to := range tos {
					confirm(to, "IN", confirmedNum, confirmdata)
				}
			case string:
				confirm(tos, "IN", confirmedNum, confirmdata)
			}
		}
		// tx.ID = ttc.ID
		// tx.RequestId = ttc.RequestId
	}

}

func onchain(address, inout string, tx models.TransferFromChain) {
	accountIds := getSubscribeIds(address)
	notifyData := map[string]interface{}{
		"status":  "TRANSFER_FROM_CHAIN",
		"inout":   inout,
		"address": address,
		"record": map[string]interface{}{
			"tfcId":     tx.ID.Hex(),
			"chain":     tx.Chain,
			"coin":      tx.Coin,
			"tokenKey":  tx.TokenKey,
			"fee":       "0",
			"from":      tx.From,
			"to":        tx.To,
			"value":     tx.Value,
			"txid":      tx.Txid,
			"blockNum":  tx.BlockHeight,
			"blockTime": tx.BlockTime,
		},
	}
	if inout == "OUT" {
		notifyData["requestId"] = tx.RequestId
	}
	if confirmedNumber == 0 {
		notifyData["finish"] = true
	}
	for _, acc := range accountIds {
		sendNotify("TRANSFER_ACTION", acc, address, notifyData)
	}

}

func confirm(address, inout string, confirmedNum interface{}, tx map[string]interface{}) {

	accountIds := getSubscribeIds(address)

	notifyData := map[string]interface{}{
		"status":       "TRANSFER_CONFIRM",
		"inout":        inout,
		"address":      address,
		"tfcId":        tx["id"],
		"txid":         tx["txid"],
		"confirmedNum": confirmedNum,
	}
	if inout == "OUT" {
		notifyData["requestId"] = tx["requestId"]
	}
	for _, acc := range accountIds {
		sendNotify("TRANSFER_ACTION", acc, address, notifyData)
	}
}

func finish(address, inout string, tx map[string]interface{}) {

	accountIds := getSubscribeIds(address)

	notifyData := map[string]interface{}{
		"status":  "TRANSFER_FINISH",
		"inout":   inout,
		"address": address,
		"tfcId":   tx["id"],
		"txid":    tx["txid"],
	}
	if inout == "OUT" {
		notifyData["requestId"] = tx["requestId"]
	}
	for _, acc := range accountIds {
		sendNotify("TRANSFER_ACTION", acc, address, notifyData)
	}
}

func sendNotify(key, accountID, address string, data interface{}) {
	// mongo 取出来的有时会有“ " ”
	accountID = strings.Trim(accountID, "\"")
	// op := options.FindOneAndUpdate().SetUpsert(true)
	_, err := db.GetCollection(commondb, "notifytasks").InsertOne(context.Background(), bson.M{"key": key, "data": data, "address": address, "_account": accountID})

	if err != nil {
		logger.Error("通知 消息 存库失败", err)
	}
}

func addSubscribesHandle(address, account string) {
	id, _ := primitive.ObjectIDFromHex(account)
	where := bson.M{"_id": id}
	up := bson.M{"$addToSet": bson.M{"addresses": address}}
	db.GetCollection(commondb, "subscribes").FindOneAndUpdate(context.Background(), where, up)
}

func getSubscribeIds(address string) []string {

	cursor, err := db.GetCollection(commondb, "subscribes").Find(context.Background(), bson.M{"addresses": address})
	if err != nil {
		logger.Error("getSubscribeIds error", err)
	}
	accountID := []string{}
	defer cursor.Close(context.Background())
	if cursor.Next(context.Background()) {
		id := cursor.Current.Lookup("_id").String() //TODO： 验证
		accountID = append(accountID, id)
	}
	return accountID
}
