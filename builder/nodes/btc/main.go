package main

import (
	"century/oasis/builder/nodes/btc/jrpc"
	"century/oasis/builder/nodes/btc/models"
	"century/oasis/builder/nodes/util"

	"github.com/shopspring/decimal"

	"context"
	"flag"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/buger/jsonparser"
	jsoniter "github.com/json-iterator/go"
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
	assign             = "TOKEN.ASSIGN"
	commondb           = "dynasty"
	chaindb            = "btc"
	chainSymbol        = "BTC"
	confirmedNumber    int64
	logger             *zap.SugaredLogger
)

// Result 返回结果
type Result struct {
	Code int         `json:"code"`
	Data interface{} `json:"data"`
	Msg  string      `json:"msg"`
}

func main() {

	flag.StringVar(&chain, "chain", "BTC", "chain")
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
		fmt.Println("connect mongo error", err)
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
	chainConf.SetFeeUrl(viper.GetString("chain.fee"))
	// // 允许来自所有域名请求
	// r.Header.Add("Access-Control-Allow-Origin", "*")
	// // 设置所允许的HTTP请求方法
	// r.Header.Add("Access-Control-Allow-Methods", "OPTIONS, GET, PUT, POST, DELETE")
	// // 字段是必需的。它也是一个逗号分隔的字符串，表明服务器支持的所有头信息字段.
	// r.Header.Add("Access-Control-Allow-Headers", "x-requested-with, accept, origin, content-type")
	// r.Header.Add("Content-Type", "application/json")

	http.HandleFunc("/api/v1/createTransferTxData", createTransactionDataHandler)
	http.HandleFunc("/api/v1/submitTxData", submitTxDtaHandler)
	http.HandleFunc("/api/v1/getBlockHeight", getBlockHeight) //getBlockInfo  getBlockHeight
	http.HandleFunc("/api/v1/getBalance", getBalance)
	http.HandleFunc("/api/v1/history", getHistory)

	url := fmt.Sprintf("%s:%s", viper.GetString("api.host"), viper.GetString("api.port"))

	fmt.Println("listen api:", url, "........")
	err := http.ListenAndServe(url, nil)
	if err != nil {
		fmt.Println("http listen failed.", err)
	}
}

func JSON(w http.ResponseWriter, data interface{}) error {
	w.Header().Set("Content-Type", "application/json")
	encoder := json.NewEncoder(w)
	return encoder.Encode(data)
}

func has(s []string, value string) bool {
	for _, v := range s {
		if v == value {
			return true
		}
	}
	return false
}
func escapeString(str, e string) {
	strings.Replace(str, "+", "%2B", -1)
}

func createTransactionDataHandler(w http.ResponseWriter, r *http.Request) {
	// fmt.Println("createTransactionDataHandler------")
	// type transfer struct {
	// 	Chain  string
	// 	Coin   string
	// 	From   string `json:"from"`
	// 	To     string
	// 	Value  string
	// 	Amount *big.Int
	// 	// TokenKey  string
	// 	RequestID string `json:"requestId"`
	// }
	// res := &Result{}
	// defer func(res *Result) {
	// 	w.Header().Add("Content-Type", "application/json")
	// 	ba, err := json.Marshal(res)
	// 	if err != nil {
	// 		res.Code = 40000
	// 		res.Msg = err.Error()
	// 	}
	// 	w.Write(ba)
	// }(res)

	// if r.Method == http.MethodPost {
	// 	fmt.Println("r.Header", r.Header)
	// 	tf := &transfer{}

	// 	fmt.Println(`r.Header.Get("Content-Type")`, r.Header.Get("Content-Type"))
	// 	switch r.Header.Get("Content-Type") {
	// 	case "application/json":
	// 		body, _ := ioutil.ReadAll(r.Body)
	// 		err := json.Unmarshal(body, tf)
	// 		if err != nil {
	// 			res.Code = 40000
	// 			res.Msg = err.Error()
	// 			// ba, _ := json.Marshal(res)
	// 			// w.Write(ba)
	// 			return
	// 		}
	// 		// str, _ := jsonparser.GetString(body, "from")
	// 		// fmt.Println("from", str)
	// 	case "application/x-www-form-urlencoded":
	// 		tf.From = r.PostFormValue("from")
	// 		tf.To = r.PostFormValue("to")
	// 		tf.TokenKey = r.PostFormValue("tokenKey")
	// 		tf.RequestID = r.PostFormValue("requestId")
	// 		tf.Value = r.PostFormValue("value")

	// 	default:
	// 		w.WriteHeader(406)
	// 		res.Code = 406
	// 		res.Msg = "not support Content-Type"
	// 		return

	// 	}
	// 	amount, ok := big.NewInt(0).SetString(tf.Value, 0)

	// 	if !ok || (amount.IsUint64() && amount.Uint64() == 0) {
	// 		res.Code = 40000
	// 		res.Msg = "Invalid amount"
	// 		return
	// 	}
	// 	tf.Amount = amount

	// 	tf.Coin = chainSymbol
	// 		tf.TokenKey = "-"

	// 	_, err := db.GetCollection(commondb, "transfertochains").InsertOne(context.Background(),
	// 		bson.M{"chain": "VCT", "coin": tf.Coin, "from": tf.From, "to": tf.To, "tokenKey": tf.TokenKey, "value": tf.Value, "requestId": tf.RequestID})

	// 	if err != nil {
	// 		fmt.Println("InsertOne transfertochains error", err)
	// 	}
	// 	// fmt.Println("insertresult", insertresult)

	// 	// res, err := chainConf.CreateTransactionData(from, to, tokenKey, amount)
	// 	_, err = chainConf.CreateTransactionData(tf.From, tf.To, tf.TokenKey, tf.Amount)
	// 	if err != nil {
	// 		res.Code = 40000
	// 		res.Msg = err.Error()
	// 		// ba, _ := json.Marshal(res)
	// 		// w.Write(ba)
	// 		return
	// 	}
	// 	resp, err := chainConf.ToResponse()
	// 	if err != nil {
	// 		res.Code = 40000
	// 		res.Msg = err.Error()
	// 		// ba, _ := json.Marshal(res)
	// 		// w.Write(ba)
	// 		return
	// 	}
	// 	res.Code = 0
	// 	res.Data = map[interface{}]interface{}{
	// 		"txData": resp.Result,
	// 	}
	// 	fmt.Println("res", res)

	// 	// w.Header().Add("Content-Type", "application/json")
	// 	// ba, _ := json.Marshal(res)
	// 	// w.Write(ba)
	// 	// fmt.Fprintln(w, string(ba))
	// 	return
	// }
	// w.WriteHeader(405)
	// res.Code = 405
	// res.Msg = "method not allow"
	// // ba, _ := json.Marshal(res)
	// // w.Write(ba)
	// return
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

	address := r.FormValue("address")
	fmt.Println("-------------- getBalance", address)
	if address == "" {
		res.Code = 40000
		res.Msg = "address empty"

		return
	}
	where := bson.D{{"address", address}}

	if currentBlockNumber > 0 {
		where = bson.D{{"address", address}, {"blockHeight", bson.M{"$lt": currentBlockNumber}}}
	}
	result, _ := db.GetCollection(chaindb, "utxos").Find(context.Background(), where)

	b := 0

	for result.Next(context.Background()) {
		var utxo struct {
			Value int `bson:"value"` // TODO:小数？
		}
		result.Decode(&utxo)
		b = b + utxo.Value
	}

	res.Code = 0
	res.Data = map[interface{}]interface{}{
		"total": b,
	}
	return

	// ba, _ := json.Marshal(res)
	// address := r.FormValue("address")
	// fmt.Println("-------------- getBalance", address)
	// if address == "" {
	// 	res := &Result{
	// 		Code: 40000,
	// 		Msg:  "address empty",
	// 	}
	// 	ba, _ := json.Marshal(res)
	// 	w.Write(ba)
	// 	return
	// }
	// b, err := chainConf.GetBalance(address)
	// if err != nil {
	// 	res := &Result{
	// 		Code: 40000,
	// 		Msg:  err.Error(),
	// 	}
	// 	ba, _ := json.Marshal(res)
	// 	w.Write(ba)
	// 	return
	// }

	// res := &Result{
	// 	Code: 0,
	// 	Data: map[interface{}]interface{}{
	// 		"total": b,
	// 	},
	// }

	// ba, _ := json.Marshal(res)
	// // w.Write(ba)

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
	fmt.Println("-+++++++++", data)
	tfcs := []models.TransferFromChain{}
	h, _ := strconv.ParseInt(data, 10, 0)
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
			fmt.Println("get transferaction error", err)
		}
		tfc := models.TransferFromChain{
			Chain:       chainSymbol,
			Coin:        chainSymbol,
			TokenKey:    "-",
			BlockHeight: tx.BlockHeight,
			BlockTime:   tx.BlockTime,
			Txid:        tx.Txid,
			Vins:        tx.Vins,
			Vouts:       tx.Vouts,
		}

		tos := []string{}
		froms := []string{}
		for _, vins := range tx.Vins {
			intxid := vins["txid"]
			vout := vins["vout"]
			where := bson.D{{"txid", intxid}, {"vout", vout}}
			var ins map[string]interface{}
			db.GetCollection(chaindb, "utxos").FindOne(context.Background(), where).Decode(&ins)
			if address := getAddressByUTXO(ins); address != "" {
				froms = append(froms, address)
			}
		}

		total := decimal.Zero
		for _, vouts := range tx.Vouts {
			v := vouts["value"]
			vv := v.(float64)
			total = total.Add(decimal.NewFromFloat(vv))
			if address := getAddressByUTXO(vouts); address != "" {
				tos = append(tos, address)
			}
		}

		tfc.From = froms
		tfc.To = tos
		if len(froms) == 0 {
			tfc.From = "coinbase"
		}

		tfc.Value = total.String()

		tfcs = append(tfcs, tfc)
	}
	return tfcs
}

func getAddressByUTXO(vouts map[string]interface{}) string {

	script := vouts["scriptPubKey"]
	sc, _ := json.Marshal(script)
	address, err := jsonparser.GetString(sc, "addresses", "[0]")
	if err != nil && err != jsonparser.KeyPathNotFoundError {
		logger.Error("jsonparser.GetString vout-->scriptPubKey--->addresses [0] error", err)
	} else {
		return address
	}
	return ""
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
		fmt.Println("transfer not found!")
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
		fmt.Println("error", tc.Err())
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
		fmt.Println("transfer not found!")
		return
	}

	// TODO: 时间类型
	// 更新转账操作记录
	updateStr := bson.M{"$set": bson.M{"code": -1, "status": `ERROR`, "updatedAt": time.Now().Unix()}, "$push": bson.M{"logs": `TX_ERROR at: ` + time.Now().String() + msg}}
	transfer := db.GetCollection(commondb, "transfers").FindOneAndUpdate(ctx, where, updateStr)

	// 更新转账账单记录
	tc := db.GetCollection(chaindb, "transferstochains").FindOneAndDelete(ctx, bson.M{"requestId": requestID})

	if tc.Err() != nil {
		fmt.Println("error", tc.Err())
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

	switch froms := tx.From.(type) {
	case []string:
		for _, from := range froms {
			onchain(from, "OUT", tx)
		}
	case string:
		onchain(froms, "OUT", tx)
	}

	switch tos := tx.To.(type) {
	case []string:
		for _, to := range tos {
			onchain(to, "OUT", tx)
		}
	case string:
		onchain(tos, "OUT", tx)
	}
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
		fmt.Println("transfer not found!")
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
					finish(to, "OUT", confirmdata)
				}
			case string:
				finish(tos, "OUT", confirmdata)
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
					confirm(to, "OUT", confirmedNum, confirmdata)
				}
			case string:
				confirm(tos, "OUT", confirmedNum, confirmdata)
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
	// fmt.Println("sendNotify", key, accountID, address, data)
	// mongo 取出来的有时会有“ " ”
	accountID = strings.Trim(accountID, "\"")
	// op := options.FindOneAndUpdate().SetUpsert(true)
	_, err := db.GetCollection(commondb, "notifytasks").InsertOne(context.Background(), bson.M{"key": key, "data": data, "address": address, "_account": accountID})

	if err != nil {
		logger.Error("通知 消息 存库失败", err)
	}
	// fmt.Println("insertresult", insertresult)
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
