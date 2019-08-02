package main

import (
	"century/oasis/builder/nodes/vct/util/chain"
	"century/oasis/builder/nodes/vct/util/comm"
	"century/oasis/builder/nodes/vct/util/dbs"
	"century/oasis/builder/nodes/vct/util/dbs/models"
	"context"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"math/big"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/buger/jsonparser"
	"github.com/json-iterator/go"
	"github.com/spf13/viper"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var (
	chain              string
	env                string
	json               = jsoniter.ConfigCompatibleWithStandardLibrary
	db                 dbs.MongoI
	currentBlockNumber = 0
	chainConf          gchain.ChainApi
	kModel             comm.KInterface
	assign             = "TOKEN.ASSIGN"
	commdb             = "dynasty"
	chaindb            = "vct"
)

// Result 返回结果
type Result struct {
	Code int         `json:"code"`
	Data interface{} `json:"data"`
	Msg  string      `json:"msg"`
}

func main() {

	flag.StringVar(&chain, "chain", "VCT", "chain")
	flag.StringVar(&env, "env", "dev", "env")
	flag.Parse()

	// viper.SetConfigFile("")
	gopath := os.Getenv("GOPATH")

	for _, p := range filepath.SplitList(gopath) {
		path := filepath.Join(p, "src/century/oasis/builder/config", strings.ToLower(env), "nodes")
		// viper.AddConfigPath(path)
		InitViper(strings.ToLower(chain), strings.ToLower(chain), path)
	}

	db = dbs.New(viper.GetString("db.addr"))
	err := db.GetConn()
	if err != nil {
		fmt.Println("connect mongo error", err)
	}

	kModel = comm.NewConsumer(viper.GetStringSlice("kafka.service"))
	// if kModel.Consumer == nil {
	// 	fmt.Println("init kafka fail-------")
	// }
	defer kModel.Close()
	go loopReadAndPaser()
	nodeURL := fmt.Sprintf("%s:%s", viper.GetString("service.host"), viper.GetString("service.port"))
	fmt.Println("nodeURL", nodeURL)
	router(nodeURL)

}

func router(url string) {
	api := fmt.Sprintf("%s://%s:%s", viper.GetString("node.protocal"), viper.GetString("node.host"), viper.GetString("node.port"))
	chainConf = gchain.NewChainAPi(api)

	// // 允许来自所有域名请求
	// r.Header.Add("Access-Control-Allow-Origin", "*")
	// // 设置所允许的HTTP请求方法
	// r.Header.Add("Access-Control-Allow-Methods", "OPTIONS, GET, PUT, POST, DELETE")
	// // 字段是必需的。它也是一个逗号分隔的字符串，表明服务器支持的所有头信息字段.
	// r.Header.Add("Access-Control-Allow-Headers", "x-requested-with, accept, origin, content-type")
	// r.Header.Add("Content-Type", "application/json")
	// r.Header.Add("Content-Type", "application/json")

	http.HandleFunc("/api/v1/createTransferTxData", createTransactionDataHandler)
	http.HandleFunc("/api/v1/submitTxData", submitTxDtaHandler)
	http.HandleFunc("/api/v1/getBlockHeight", getBlockHeight)
	http.HandleFunc("/api/v1/getBalance", getBalance)

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
	fmt.Println("createTransactionDataHandler------")
	type transfer struct {
		Chain     string
		Coin      string
		From      string `json:"from"`
		To        string
		Value     string
		Amount    *big.Int
		TokenKey  string
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
		fmt.Println("r.Header", r.Header)
		tf := &transfer{}

		fmt.Println(`r.Header.Get("Content-Type")`, r.Header.Get("Content-Type"))
		switch r.Header.Get("Content-Type") {
		case "application/json":
			body, _ := ioutil.ReadAll(r.Body)
			err := json.Unmarshal(body, tf)
			if err != nil {
				res.Code = 40000
				res.Msg = err.Error()
				// ba, _ := json.Marshal(res)
				// w.Write(ba)
				return
			}
			// str, _ := jsonparser.GetString(body, "from")
			// fmt.Println("from", str)
		case "application/x-www-form-urlencoded":
			tf.From = r.PostFormValue("from")
			tf.To = r.PostFormValue("to")
			tf.TokenKey = r.PostFormValue("tokenKey")
			tf.RequestID = r.PostFormValue("requestId")
			tf.Value = r.PostFormValue("value")

		default:
			w.WriteHeader(406)
			res.Code = 406
			res.Msg = "not support Content-Type"
			// ba, _ := json.Marshal(res)
			// w.Write(ba)
			return

		}
		amount, ok := big.NewInt(0).SetString(tf.Value, 0)

		if !ok || (amount.IsUint64() && amount.Uint64() == 0) {
			// s.NormalErrorF(rw, 0, "Invalid amount")
			res.Code = 40000
			res.Msg = "Invalid amount"
			// ba, _ := json.Marshal(res)
			// w.Write(ba)
			return
		}
		tf.Amount = amount

		if tf.TokenKey == "" || tf.TokenKey == "-" {
			tf.Coin = "VCT"
			tf.TokenKey = "-"
		} else {
			tf.Coin = "VCT_TOKEN"
		}

		insertresult, err := db.GetCollection(commdb, "transfertochains").InsertOne(context.Background(),
			bson.M{"chain": "VCT", "coin": tf.Coin, "from": tf.From, "to": tf.To, "tokenKey": tf.TokenKey, "value": tf.Value, "requestId": tf.RequestID})

		if err != nil {
			fmt.Println("InsertOne transfertochains error", err)
		}
		fmt.Println("insertresult", insertresult)

		// res, err := chainConf.CreateTransactionData(from, to, tokenKey, amount)
		_, err = chainConf.CreateTransactionData(tf.From, tf.To, tf.TokenKey, tf.Amount)
		if err != nil {
			res.Code = 40000
			res.Msg = err.Error()
			// ba, _ := json.Marshal(res)
			// w.Write(ba)
			return
		}
		resp, err := chainConf.ToResponse()
		if err != nil {
			res.Code = 40000
			res.Msg = err.Error()
			// ba, _ := json.Marshal(res)
			// w.Write(ba)
			return
		}
		res.Code = 0
		res.Data = map[interface{}]interface{}{
			"txData": resp.Result,
		}
		fmt.Println("res", res)

		// w.Header().Add("Content-Type", "application/json")
		// ba, _ := json.Marshal(res)
		// w.Write(ba)
		// fmt.Fprintln(w, string(ba))
		return
	}
	w.WriteHeader(405)
	res.Code = 405
	res.Msg = "method not allow"
	// ba, _ := json.Marshal(res)
	// w.Write(ba)
	return
}

func submitTxDtaHandler(w http.ResponseWriter, r *http.Request) {
	requestID := r.PostFormValue("requestId")
	singedRawTx := r.PostFormValue("singedRawTx")

	res := &Result{}
	defer func(res *Result) {
		if res.Code != 0 {
			setSendTransactionError(requestID, res.Msg)
		}

		ba, _ := json.Marshal(res)
		w.Header().Add("Content-Type", "application/json")
		w.Write(ba)
	}(res)

	id, _ := primitive.ObjectIDFromHex(requestID) // requestID 需要转化为objectId
	where := bson.M{"_id": id}                    //insertresult.InsertedID
	result := db.GetCollection(commdb, "transfers").FindOne(context.Background(), where)

	// type singedTx struct {
	// 	TxData map[string]string `json:"txData"`
	// }
	// var tx = &singedTx{}
	// result.Decode(tx)

	rawByte, _ := result.DecodeBytes()
	raw := rawByte.Lookup("txData", "raw").String() // 坑 返回的是json ，单纯字符串会有 /""/ 应去掉
	raw = strings.Trim(raw, "\"")

	if raw == "" {
		res.Code = 40000
		res.Msg = "not find raw tx"
		// ba, _ := json.Marshal(res)
		// w.Write(ba)
		return
	}

	b, err := chainConf.SubmitTransactionData(raw, singedRawTx)
	if err != nil {
		res.Code = 40000
		res.Msg = err.Error()
		return
	}

	resp, _ := chainConf.ToResponse()
	fmt.Println("resp", resp)

	txid, err := jsonparser.GetString(b, "result")
	if err != nil {
		res.Code = 40000
		res.Msg = err.Error()
		return
	}
	fmt.Println("res", txid)

	setSendTransactionTxid(requestID, txid)
	// data := &Result{
	// 	Code: 0,
	// 	Data: map[interface{}]interface{}{
	// 		"txid": txid,
	// 	},
	// }
	res.Code = 0
	res.Data = map[string]interface{}{
		"txid": txid,
	}
	// ba, _ := json.Marshal(data)

	// w.Write(ba)
	return
	// fmt.Fprintln(w, string(ba))
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

	fmt.Fprintln(w, string(b))
}

func getBalance(w http.ResponseWriter, r *http.Request) {

	address := r.FormValue("address")
	fmt.Println("-------------- getBalance", address)
	if address == "" {
		res := &Result{
			Code: 40000,
			Msg:  "address empty",
		}
		ba, _ := json.Marshal(res)
		w.Write(ba)
		return
	}
	b, err := chainConf.GetBalance(address)
	if err != nil {
		res := &Result{
			Code: 40000,
			Msg:  err.Error(),
		}
		ba, _ := json.Marshal(res)
		w.Write(ba)
		return
	}

	res := &Result{
		Code: 0,
		Data: map[interface{}]interface{}{
			"total": b,
		},
	}

	ba, _ := json.Marshal(res)
	// w.Write(ba)

	// // 允许来自所有域名请求
	w.Header().Add("Access-Control-Allow-Origin", "*")
	// 设置所允许的HTTP请求方法
	w.Header().Add("Access-Control-Allow-Methods", "OPTIONS, GET, PUT, POST, DELETE")
	// 字段是必需的。它也是一个逗号分隔的字符串，表明服务器支持的所有头信息字段.
	w.Header().Add("Access-Control-Allow-Headers", "x-requested-with, accept, origin, content-type")
	w.Header().Add("Content-Type", "application/json")
	// fmt.Fprintln(w, string(ba))
	fmt.Println("ba", ba)
	w.Write(ba)
}

//InitViper we can set viper which fabric peer is used
func InitViper(envprefix string, filename string, configPath ...string) error {
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

func loopReadAndPaser() {
	msg := make(chan []byte)
	go func() {
		kModel.ReciveMsg(msg)
	}()

	for {
		select {

		case m := <-msg:
			tfcs := paserTx(m)
			for _, tfc := range tfcs {
				newTranferFromChain(tfc, false)
			}
		}
	}

	// kModel.ReciveMsg()
}

func paserTx(msg []byte) []models.TransferFromChain {
	fmt.Println("-+++++++++", string(msg))
	tfcs := []models.TransferFromChain{}
	h := string(msg)
	// fmt.Println("-+++++++++", string(blockInfo))
	type res struct {
		Result *models.Blocks `json:"result"`
	}
	// fmt.Println("--------------- blockInfo", string(blockInfos))
	curor, err := db.GetCollection("vct", "transactions").Find(context.Background(), bson.M{"blockheight": h})

	if err != nil {
		fmt.Println("get transferaction error")
	}

	defer curor.Close(context.Background())
	for curor.Next(context.Background()) {
		tx := models.Transaction{}
		if err := curor.Decode(&tx); err != nil {
			fmt.Println("get transferaction error", err)
		}
		tfc := models.TransferFromChain{
			Chain:    "VCT",
			Coin:     "VCT",
			TokenKey: "-",
		}
		log.Println(tx)
		// op := options.FindOneAndUpdate().SetUpsert(true)

		if tx.OnChain {
			tfc.BlockHeight = tx.BlockHeight
			tfc.BlockTime = tx.BlockTime
			tfc.From = tx.From
			tfc.To = tx.To
			tfc.Txid = tx.Txid
			tfc.Value = tx.Value
			if tx.TokenKey != "" && tx.TokenKey != "-" {
				tfc.TokenKey = tx.TokenKey
				tfc.Coin = "VCT_TOKEN"
			}

		} else {

		}
		tfcs = append(tfcs, tfc)

		// if tx.BlockHeight != "" {
		// 	tx.Chain = "VCT"
		// 	tx.Coin = "VCT_TOKEN"
		// 	op := options.FindOneAndUpdate().SetUpsert(true)

		// 	db.GetCollection("vct", "transferfromchains").FindOneAndUpdate(context.Background(), bson.M{"txid": tx.Txid}, bson.M{"$set": tx}, op)
		// }
	}
	return tfcs
}

func setSendTransactionTxid(requestID, txid string) {
	// const transfer = await this.commdb.models.Transfer.findById(mid)
	id, _ := primitive.ObjectIDFromHex(requestID)
	where := bson.M{"_id": id}
	ctx := context.Background()
	result := db.GetCollection(commdb, "transfers").FindOne(ctx, where)
	if result.Err() != nil {
		fmt.Println("%s", "transfer not found!")
	}
	// 更新转账操作记录
	updateStr := bson.M{"$set": bson.M{"txid": txid, "code": 16, "status": `TXID`, "updatedAt": time.Now().Unix()}, "$push": bson.M{"logs": `TX_HASH at: ${Date.now()}`}}
	db.GetCollection(commdb, "transfers").FindOneAndUpdate(ctx, where, updateStr)
	// // 更新转账账单记录
	updateStr1 := bson.M{"$set": bson.M{"code": 16, "txid": txid, "updatedAt": time.Now().Unix()}}
	tc := db.GetCollection(chaindb, "transferstochains").FindOneAndUpdate(ctx, bson.M{"requestId": requestID}, updateStr1)

	if tc.Err() != nil {
		fmt.Println("error", tc.Err())
	}
	tcdecode := models.TransferToChain{}
	tc.Decode(&tcdecode)

	// // 构造消息
	// const { _account } = transfer
	// const notifyData = {
	// 	status: 'SUBMIT_TRANSACTION_TO_CHAIN',
	// 	description: 'submit transfer transaction to chain',
	// 	requestId: mid,
	// 	tfcId: tc._id,
	// 	txid,
	// }
	// await this.sendNotify('TRANSFER_ACTION', notifyData, _account)
	notifyData := map[string]string{
		"status":    "SUBMIT_TRANSACTION_TO_CHAIN",
		"requestId": requestID,
		"tfcId":     tcdecode.ID.String(),
		"txid":      txid,
	}
	sendNotify("TRANSFER_ACTION", tcdecode.From, "", notifyData) //TODO : _account
}

func setSendTransactionError(requestID, msg string) {
	// const transfer = await this.commdb.models.Transfer.findById(mid)
	commDb := "dynasty"
	chaindb := "vct"
	id, _ := primitive.ObjectIDFromHex(requestID)
	where := bson.M{"_id": id}
	ctx := context.Background()
	result := db.GetCollection(chaindb, "transferstochain").FindOneAndDelete(ctx, bson.M{"requestID": requestID})
	if result.Err() != nil {
		fmt.Println("transfer not found!")
	}
	// if !transfer {
	// 	fmt.Errorf("%s \n", "transfer not found!")
	// 	return
	// }
	// 更新转账操作记录
	updateStr := bson.M{"$set": bson.M{"code": -1, "status": `ERROR`, "updatedAt": time.Now().Unix()}, "$push": bson.M{"logs": `TX_ERROR at: ${Date.now()}` + msg}}
	db.GetCollection(commDb, "transfers").FindOneAndUpdate(ctx, where, updateStr)
	// // 更新转账账单记录
	// const updateStr1 = { $set: { code: 16, txid, updatedAt: Date.now() } }
	// const tc = await this.chaindb.models.TransferToChain.findOneAndUpdate({ requestId: mid }, updateStr1)
	// // 更新转账账单记录
	tc := db.GetCollection(chaindb, "transferstochains").FindOneAndDelete(ctx, bson.M{"requestId": requestID})

	if tc.Err() != nil {
		fmt.Println("error", tc.Err())
	}
	tcdecode := models.TransferToChain{}
	tc.Decode(&tcdecode)
	// // 构造消息
	// const { _account } = transfer
	// const notifyData = {
	// 	status: 'SUBMIT_TRANSACTION_TO_CHAIN',
	// 	description: 'submit transfer transaction to chain',
	// 	requestId: mid,
	// 	tfcId: tc._id,
	// 	txid,
	// }
	// await this.sendNotify('TRANSFER_ACTION', notifyData, _account)
	notifyData := map[string]string{
		"status":    "SUBMIT_TRANSACTION_ERROR",
		"requestId": requestID,
		"tfcId":     tcdecode.ID.String(),
		// "txid":      txid,
		"error": msg,
	}

	sendNotify("TRANSFER_ACTION", "", tcdecode.From, notifyData) //TODO: _account
}

func newTranferFromChain(tfc models.TransferFromChain, haveComfirming bool) {
	tx := tfc
	dbname := tx.Chain
	commdb = "dynasty"
	tx.CreatedAt = time.Now().Unix()
	op := options.FindOneAndUpdate().SetUpsert(true)
	ctx := context.Background()
	where := bson.M{"txid": tx.Txid}
	ttcResult := db.GetCollection(dbname, "transferTochains").FindOne(ctx, where)

	var updateStr bson.M
	if ttcResult != nil && ttcResult.Err() == nil {
		ttc := models.TransferToChain{}
		ttcResult.Decode(&ttc)
		tx.ID = ttc.ID
		tx.RequestId = ttc.RequestId

		updateStr1 := bson.M{"$set": bson.M{"code": 32, "status": "FROM_CHAIN", "updatedAt": time.Now().Unix()}, "$push": bson.M{"logs": `FROM_CHAIN at: ${Date.now()}`}}

		id, _ := primitive.ObjectIDFromHex(ttc.RequestId)
		db.GetCollection(commdb, "transfers").FindOneAndUpdate(ctx, bson.M{"_id": id}, updateStr1)

		// 添加地址到订阅

		addSubscribesHandle(ttc.From, ttc.ID.String())
	}
	updateStr = bson.M{"$set": tx}

	if !haveComfirming {
		db.GetCollection(dbname, "transferfromchains").FindOneAndUpdate(context.Background(), bson.M{"txid": tx.Txid}, updateStr, op)
	} else {
		// db.GetCollection(dbname, "transferfromchains").FindOneAndUpdate(context.Background(), bson.M{"txid": tx.Txid}, updateStr, op)
	}

	// TODO: 查询 订阅表
	onchain(tx.From, "OUT", tx)
	onchain(tx.To, "IN", tx)

}

func newBlockNotify(blockNumber string) {
	// if (this.inProcess) return

	// this.inProcess = true
	chaindb := "vct"
	// commdb := "dynasty"
	ctx := context.Background()
	// const safeBlockNumber = blockNumber - this.confirmedMaxNum + 1
	// // this.logger.info('[DynastyThreadUtil:newBlockNotify]', blockNumber, 'safeBlockNumber', safeBlockNumber)
	// const tcs = await this.chaindb.models.TransferConfirming.find().limit(10240)
	op := options.Find().SetLimit(1024)
	result, err := db.GetCollection(chaindb, "transferconfirmings").Find(ctx, bson.M{}, op)
	if err != nil {
		fmt.Println("transfer not found!")
	}
	if result.Next(ctx) {

	}

	// 	const ts = Date.now()

	// 	for (const tc of tcs) {
	// 		try {
	// 			const { blockHeight, requestId, from, to } = tc
	// 			// 确认完毕
	// 			const isFinish = blockHeight <= safeBlockNumber
	// 			const confirmedNum = blockNumber - blockHeight + 1

	// 			// console.log(tc.toObject(), isFinish, blockHeight, safeBlockNumber)

	// 			if (isFinish) {
	// 				// 1. 创建TransferFromChain
	// 				let tfc
	// 				try {
	// 					// "$isolated": true
	// 					tfc = await this.chaindb.models.TransferFromChain.findByIdAndUpdate(tc._id, { $set: tc.toObject() }, { upsert: true, new: true })
	// 				} catch (e) {
	// 					if (e.message.indexOf('E11000') >= 0) {
	// 						this.logger.warn(e.message)
	// 					}
	// 				}
	// 				// 2. 修改转账记录状态
	// 				if (requestId) {
	// 					await this.commdb.models.Transfer.findByIdAndUpdate(requestId, { $set: { code: 64, status: 'FINISH', updatedAt: ts }, $push: { logs: `confirm finish at ${ts}` } })
	// 				}
	// 				await this.chaindb.models.TransferConfirming.findByIdAndDelete(tc._id)
	// 				const froms = Array.isArray(from) ? from : [from]
	// 				for (const addr of froms) {
	// 					const fromSubscribed = this.addressMap[addr]
	// 					if (fromSubscribed) {
	// 						await this._finish(fromSubscribed, addr, tfc, 'OUT')
	// 					}
	// 				}
	// 				const tos = Array.isArray(to) ? to : [to]
	// 				for (const addr of tos) {
	// 					const toSubscribed = this.addressMap[addr]
	// 					if (toSubscribed) {
	// 						console.log('__toall finish(IN):', addr)
	// 						await this._finish(toSubscribed, addr, tfc, 'IN')
	// 					}
	// 				}

	// 				// 一笔交易, 钱包里还会包含一笔到from的找零
	// 				if ((this.chainSymbol === 'BTC' || this.chainSymbol === 'LTC') && tc.outs && tc.outs.length > 0) {
	// 					for (const out of tc.outs) {
	// 						const o = tos.find(item => String(item) === String(out.address))
	// 						if (!o || o.length === 0) {
	// 							const toSubscribed = this.addressMap[out.address]
	// 							if (toSubscribed) {
	// 								console.log('__toall finish(IN)(From):', out.address)
	// 								await this._finish(toSubscribed, out.address, tfc, 'IN')
	// 							}
	// 						}
	// 					}
	// 				}

	// 				// 1. 删除待确认记录
	// 				await this.chaindb.models.TransferConfirming.findByIdAndDelete(tc._id)
	// 			} else {
	// 				const froms = Array.isArray(from) ? from : [from]
	// 				for (const addr of froms) {
	// 					const fromSubscribed = this.addressMap[addr]
	// 					if (fromSubscribed) {
	// 						this.logger.info('__toall _confirm(OUT):', confirmedNum, blockNumber)
	// 						await this._confirm(fromSubscribed, tc._id, 'OUT', confirmedNum, blockNumber)
	// 					}
	// 				}
	// 				const tos = Array.isArray(to) ? to : [to]
	// 				for (const addr of tos) {
	// 					const toSubscribed = this.addressMap[addr]
	// 					if (toSubscribed) {
	// 						this.logger.info('__toall _confirm(IN):', confirmedNum, blockNumber)
	// 						await this._confirm(toSubscribed, tc._id, 'IN', confirmedNum, blockNumber)
	// 					}
	// 				}
	// 			}
	// 		} catch (e) {
	// 			this.logger.error('[DynastyThreadUtil:newBlockNotify]', e)
	// 		}
	// 	}
	// } finally {
	// 	// this.inProcess = false
	// }
}

func onchain(address, inout string, tx models.TransferFromChain) {
	accountIds := getSubscribeIds()
	notifyData := map[string]interface{}{
		"status":  "TRANSFER_FROM_CHAIN",
		"inout":   inout,
		"address": address,
		"record": map[string]string{
			"id":        tx.ID.String(),
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
	for _, acc := range accountIds {
		sendNotify("TRANSFER_ACTION", acc, address, notifyData) //TODO: _account
	}

}

func finish(address, inout string, tx models.TransferFromChain) {

	accountIds := getSubscribeIds()

	notifyData := map[string]interface{}{
		"status":   "TRANSFER_FINISH",
		"inout":    inout,
		"address":  address,
		"id":       tx.ID.String(),
		"blockNum": tx.BlockHeight,
		"chain":    tx.Chain,
	}
	if inout == "OUT" {
		notifyData["requestId"] = tx.RequestId
	}
	for _, acc := range accountIds {
		sendNotify("TRANSFER_ACTION", acc, address, notifyData) //TODO: _account
	}
}

func sendNotify(key, accountID, address string, data interface{}) {

	// op := options.FindOneAndUpdate().SetUpsert(true)
	insertresult, err := db.GetCollection(commdb, "notifytasks").InsertOne(context.Background(), bson.M{"key": key, "data": data, "address": address, "_account": accountID})

	if err != nil {
		fmt.Println("通知 消息 存库失败", err)
	}
	fmt.Println("insertresult", insertresult)
	// const task = new this.commdb.models.NotifyTask({ key, data, _account: accountId, address })
	// const s = await task.save()
	// this.sendToKafka({ accountId, id: s._id }, 'NOTIFY_TASK', key)
}

func addSubscribesHandle(address, account string) {
	id, _ := primitive.ObjectIDFromHex(account)
	where := bson.M{"_id": id}
	up := bson.M{"$addToSet": bson.M{"addresses": address}}
	db.GetCollection(commdb, "subscribes").FindOneAndUpdate(context.Background(), where, up)
}

func getSubscribeIds(address string) []string {
	commdb := "dynasty"
	cursor, _ := db.GetCollection(commdb, "subscribes").Find(context.Background(), bson.M{"addresses": address})
	accountID := []string{}
	defer cursor.Close(context.Background())
	if cursor.Next(context.Background()) {
		id := cursor.Current.Lookup("_id").String()
		accountID = append(accountID, id)
	}
	return accountID
}
