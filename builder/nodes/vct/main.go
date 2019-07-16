package main

import (
	"century/oasis/builder/nodes/vct/util/chain"
	"century/oasis/builder/nodes/vct/util/comm"
	"century/oasis/builder/nodes/vct/util/dbs"
	"century/oasis/builder/nodes/vct/util/dbs/models"
	"context"
	"flag"
	"fmt"
	"log"
	"math/big"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/json-iterator/go"
	"github.com/spf13/viper"
	"go.mongodb.org/mongo-driver/bson"
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
)

// Result 返回结果
type Result struct {
	Code int                    `json:code`
	Data map[string]interface{} `json:Data`
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

	api := fmt.Sprintf("%s://%s:%s", viper.GetString("node.protocal"), viper.GetString("node.host"), viper.GetString("node.port"))
	chainConf = gchain.NewChainAPi(api)
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
	http.HandleFunc("/api/v1/createTransferTxData", createTransactionDataHandler)
	http.HandleFunc("/api/v1/submitTxDta", submitTxDtaHandler)
	http.HandleFunc("/api/v1/getBlockHeight", getBlockHeight)

	err := http.ListenAndServe(url, nil)
	if err != nil {
		fmt.Println("http listen failed.", err)
	}
}

func createTransactionDataHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		from := r.PostFormValue("from")
		to := r.PostFormValue("to")
		// value := r.PostFormValue("value")
		tokenKey := r.PostFormValue("tokenKey")

		amount, ok := big.NewInt(0).SetString(r.PostFormValue("value"), 0)

		if !ok || (amount.IsUint64() && amount.Uint64() == 0) {
			// s.NormalErrorF(rw, 0, "Invalid amount")
			fmt.Fprintln(w, "Invalid amount")
			return
		}
		res, _ := chainConf.CreateTransactionData(from, to, tokenKey, amount)
		fmt.Fprintln(w, res)
	} else {
		fmt.Fprintln(w, "only Post ")
	}

}

func submitTxDtaHandler(w http.ResponseWriter, r *http.Request) {

	from := r.PostFormValue("from")
	to := r.PostFormValue("to")

	tokenKey := r.PostFormValue("tokenKey")

	amount, ok := big.NewInt(0).SetString(r.PostFormValue("value"), 0)

	if !ok || (amount.IsUint64() && amount.Uint64() == 0) {
		fmt.Fprintln(w, "Invalid amount")
		return
	}
	res, _ := chainConf.CreateTransactionData(from, to, tokenKey, amount)
	fmt.Fprintln(w, res)
}

func getBlockHeight(w http.ResponseWriter, r *http.Request) {
	fmt.Println("-------------- getBlockHeight")
	h, err := chainConf.GetBlockHeight()
	if err != nil {
		fmt.Println("--------------err", err)
	}

	res := &Result{
		Code: 0,
		Data: map[string]interface{}{
			"Height": h,
		},
	}

	b, _ := json.Marshal(res)

	fmt.Fprintln(w, string(b))
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
	commDb := "dynasty"
	where := bson.M{"_id": requestID}
	ctx := context.Background()
	result := db.GetCollection(commDb, "transfers").FindOne(ctx, where)
	if result.Err() != nil {
		fmt.Errorf("%s \n", "transfer not found!")
	}
	// if !transfer {
	// 	fmt.Errorf("%s \n", "transfer not found!")
	// 	return
	// }
	// 更新转账操作记录
	updateStr := bson.M{"$set": bson.M{"txid": txid, "code": 16, "status": `TXID`, "updatedAt": time.Now().Unix()}, "$push": bson.M{"logs": `TX_HASH at: ${Date.now()}`}}
	db.GetCollection(commDb, "transfers").FindOneAndUpdate(ctx, requestID, updateStr)
	// // 更新转账账单记录
	// const updateStr1 = { $set: { code: 16, txid, updatedAt: Date.now() } }
	// const tc = await this.chaindb.models.TransferToChain.findOneAndUpdate({ requestId: mid }, updateStr1)

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
}

func setSendTransactionError(requestID, txid string) {
	// const transfer = await this.commdb.models.Transfer.findById(mid)
	commDb := "dynasty"
	chaindb := "vct"
	where := bson.M{"_id": requestID}
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
	updateStr := bson.M{"$set": bson.M{"txid": txid, "code": 16, "status": `TXID`, "updatedAt": time.Now().Unix()}, "$push": bson.M{"logs": `TX_HASH at: ${Date.now()}`}}
	db.GetCollection(commDb, "transfers").FindOneAndUpdate(ctx, where, updateStr)
	// // 更新转账账单记录
	// const updateStr1 = { $set: { code: 16, txid, updatedAt: Date.now() } }
	// const tc = await this.chaindb.models.TransferToChain.findOneAndUpdate({ requestId: mid }, updateStr1)

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
}

func newTranferFromChain(tfc models.TransferFromChain, haveComfirming bool) {
	tx := tfc
	dbname := tx.Chain
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
		// updateStr = bson.M{"$set": tx}
	}
	updateStr = bson.M{"$set": tx}

	if !haveComfirming {
		db.GetCollection(dbname, "transferfromchains").FindOneAndUpdate(context.Background(), bson.M{"txid": tx.Txid}, updateStr, op)
	}

}

func newBlockNotify(blockNumber string) {
	// if (this.inProcess) return

	// this.inProcess = true
	chaindb := "vct"
	commdb := "dynasty"
	ctx := context.Background()
	// const safeBlockNumber = blockNumber - this.confirmedMaxNum + 1
	// // this.logger.info('[DynastyThreadUtil:newBlockNotify]', blockNumber, 'safeBlockNumber', safeBlockNumber)
	// const tcs = await this.chaindb.models.TransferConfirming.find().limit(10240)
	op := options.Find().SetLimit(1024)
	result, err := db.GetCollection(chaindb, "transferconfirmings").Find(ctx, bson.M{}, op)
	if err != nil {
		fmt.Println("transfer not found!")
	}
	if result.Next() {

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
