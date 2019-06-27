package main

import (
	"century/oasis/speed/nodes/vct/util/chain"
	"century/oasis/speed/nodes/vct/util/comm"
	"century/oasis/speed/nodes/vct/util/dbs"
	"century/oasis/speed/nodes/vct/util/dbs/models"
	"context"
	"flag"
	"fmt"
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
	"go.mongodb.org/mongo-driver/mongo/options"
	// "century/oasis/speed/vct/db"
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
		path := filepath.Join(p, "src/century/oasis/speed/config", strings.ToLower(env), "nodes")
		// viper.AddConfigPath(path)
		InitViper(strings.ToLower(chain), strings.ToLower(chain), path)
	}
	// host := viper.GetString("service.host")
	// port := viper.GetString("service.port")
	// router(host + ":" + port)

	db = dbs.New(viper.GetString("db.addr"))
	err := db.GetConn()
	if err != nil {
		fmt.Println("connect mongo error", err)
	}

	api := fmt.Sprintf("%s://%s:%s", viper.GetString("node.protocal"), viper.GetString("node.host"), viper.GetString("node.port"))
	chainConf = gchain.NewChainAPi(api)
	kModel := comm.NewKafkaModel(viper.GetStringSlice("kafka.service"))
	defer kModel.Close()
	loopReadAndPaser()

}

func router(url string) {
	http.HandleFunc("/api/v1/createTransferTxDta", createTransactionDataHandler)
	http.HandleFunc("/api/v1/submitTxDta", submitTxDtaHandler)
	http.HandleFunc("/api/v1/getBlockHeight", getBlockHeight)
	// http.HandleFunc("/login/", loginHandler)
	// http.HandleFunc("/ajax/", ajaxHandler)
	// http.HandleFunc("/", NotFoundHandler)
	err := http.ListenAndServe(url, nil)
	if err != nil {
		fmt.Println("http listen failed.", err)
	}
}

func createTransactionDataHandler(w http.ResponseWriter, r *http.Request) {
	from := r.PostFormValue("from")
	to := r.PostFormValue("to")
	// value := r.PostFormValue("value")
	tokenKey := r.PostFormValue("tokenKey")

	amount, ok := big.NewInt(0).SetString(r.PostFormValue("value"), 0)

	if !ok || (amount.IsUint64() && amount.Uint64() == 0) {
		// s.NormalErrorF(rw, 0, "Invalid amount")
		fmt.Println("Invalid amount")
		return
	}
	chainConf.CreateTransactionData(from, to, tokenKey, amount)

}

func submitTxDtaHandler(w http.ResponseWriter, r *http.Request) {

	from := r.PostFormValue("from")
	to := r.PostFormValue("to")

	tokenKey := r.PostFormValue("tokenKey")

	amount, ok := big.NewInt(0).SetString(r.PostFormValue("value"), 0)

	if !ok || (amount.IsUint64() && amount.Uint64() == 0) {
		// s.NormalErrorF(rw, 0, "Invalid amount")
		fmt.Println("Invalid amount")
		return
	}
	chainConf.CreateTransactionData(from, to, tokenKey, amount)

}

func getBlockHeight(w http.ResponseWriter, r *http.Request) {

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

func initLatestBlockNumber() int64 {

	collection := db.GetCollection("vct", "infos")

	result := &models.Info{}

	filter := bson.M{}
	ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)
	err := collection.FindOne(ctx, filter).Decode(&result)
	if err != nil {
		// log.Fatal(err)
		fmt.Println(" collection.FindOne err", err)
	}

	fmt.Println("initLatestBlockNumber", result.Height)
	if result.Height > 0 {
		return result.Height + 1
	}
	return 0
}

func isNewBlockAvalible(number int64) bool {

	block, err := chainConf.GetBlockHeight()
	if err != nil {
		fmt.Println("--------------err", err)
		block = -1
	}
	return block > number

}

func getBlockInfo(number int64) []byte {

	b, err := chainConf.GetBlockInfo(number)

	if err != nil {
		return nil
	}
	return b

	// if err != nil {
	// 	fmt.Println("getBlockInfo error", err)
	// }
	// return block

}

func readAndParseBlock(number int64) {
	blockInfos := getBlockInfo(number)
	// fmt.Println("-+++++++++", string(blockInfo))
	type res struct {
		Result *models.Blocks `json:"result"`
	}
	fmt.Println("--------------- blockInfo", string(blockInfos))
	b := &res{}
	err := json.Unmarshal(blockInfos, b)
	if err != nil {
		// return -1, err
		fmt.Println("json.Unmarshal(blockInfo error", err)
	}
	fmt.Println("---------------json.Unmarshal ", b)

	h, err := jsonparser.GetString(blockInfos, "result", "Height")
	if err != nil {
		fmt.Println("jsonparser.GetString error", err)
	}
	fmt.Println("---------------jsonparser.GetString", h)

	coll := db.GetCollection("vct", "blocks")
	// err = coll.Drop(context.Background())

	if b.Result.Height != "" {
		docs := bson.M{
			"height": b.Result.Height,
			"hash":   b.Result.Hash,
			"time":   b.Result.TimeStamp,
		}

		_, err = coll.InsertOne(context.Background(), docs)
		if err != nil {
			fmt.Println("insert one err", err)
		}
		// err = coll.Drop(context.Background())

		txs := models.Transactions{
			BlockHeight: b.Result.Height,
			BlockTime:   b.Result.TimeStamp,
			BlockHash:   b.Result.Hash,
			OnChain:     true,
		}

		for _, item := range b.Result.Txs {
			fmt.Println("item", item)
			txs.TxID = item.TxID
			txs.Method = item.Method
			if item.Method == "batch" {
				jsonparser.ArrayEach(blockInfos, func(value []byte, dataType jsonparser.ValueType, offset int, err error) {
					m, err := jsonparser.GetString(value, "Method")
					if err != nil {
						fmt.Println("jsonparser.GetString Method error", err)
					}
					if m == assign {
						to, _ := jsonparser.GetString(value, "Method", "to")
						token, _ := jsonparser.GetString(value, "Method", "name")
						txs.To = to
						txs.TokenKey = token
					}

				}, "result", "Transactions", "Detail")

			} else {
				fmt.Println("item.Detail  not batch", item.Detail)
				txs.From = item.Detail.From
				txs.To = item.Detail.To
				txs.Value = item.Detail.Amount
				txs.TokenKey = item.Detail.Token
			}
		}

		for _, item := range b.Result.Events {
			// txs.From=item.
			txs.TxID = item.TxID
			txs.Log = item.Detail
			txs.OnChain = false

		}

		op := &options.FindOneAndUpdateOptions{}
		op.SetUpsert(true)
		// kModel.SendMsg("VCT_TX", string(blockInfos))
		rs := db.GetCollection("vct", "infos").FindOneAndUpdate(context.Background(), bson.M{}, bson.D{{"$set", bson.M{"height": h}}}, op)
		if rs.Err() != nil {
			fmt.Println("FindOneAndUpdate err", rs.Err())
		}

		fmt.Println("FindOneAndUpdate", rs)
		// db.GetCollection("vct", "transactions").FindOneAndUpdate(context.Background(), bson.D{}, bson.D{{"height", b.Result.Height}})
	}
	// if (blockInfo.result) {
	// 	// 交易表
	// 	// await that.db.models.Transaction.deleteMany({ blockId: height });

	// 	logger.info(`开始查询第${height} 块,删除已有数据`)

	// 	blockInfo = blockInfo.result
	// 	logger.debug(`第${height} 块数据 ${JSON.stringify(blockInfo)}`)
	// 	// blockInfo = that.testdata() // 手动构造数据测试
	// 	const rawdata = {
	// 		height: blockInfo.Height,
	// 		hash: blockInfo.Hash,
	// 		timestamp: blockInfo.TimeStamp,

	// 		transactions: blockInfo.Transactions || [],
	// 		txEvents: blockInfo.TxEvents || [],

	// 		// time: Date.parse(blockInfo.TimeStamp.split('+')[0]),

	// 		rawTime: blockInfo.TimeStamp,
	// 		time: this.dateToUnix(blockInfo.TimeStamp),
	// 	}

	// 	const txLen = rawdata.transactions.length + rawdata.txEvents.length

	// 	// logger.debug(`块 ${height} txLen ${txLen}`)

	// 	// 保存原始数据
	// 	await that.db.models.Block.findOneAndUpdate({ height }, { $set: Object.assign({}, rawdata, { txCount: txLen }) }, { upsert: true });

	// 	const { transactions, txEvents, ...baseFileds } = rawdata

	// 	// 一次读取块种50个交易记录

	// 	if (transactions && transactions.length > 0) { // 上链信息
	// 		for (const tx of transactions) {

	// 			/**
	// 			 *  {
	// 				 "Height": "4",
	// 				 "TxID": "365A858149C6E2D115868BF811B28E24",
	// 				 "Chaincode": "local",
	// 				 "Method": "batch",
	// 				 "CreatedFlag": false,
	// 				 "ChaincodeModule": "AtomicEnergy_v1",
	// 				 "Nonce": "EBCD3BC27E8F541A4215960088E6550D6A4D2D9B",
	// 				 "Detail": [
	// 					 {
	// 						 "Method": "MTOKEN.INIT",
	// 						 "Detail": "VCT31: Unknown message"
	// 					 },
	// 					 {
	// 						 "Method": "MTOKEN.ASSIGN",
	// 						 "Detail": {
	// 							 "amount": "1000000000000000000000000000",
	// 							 "to": "ARJtq6Q46oTnxDwvVqMgDtZeNxs7Ybt81A",
	// 							 "token": "VCT31"
	// 						 }
	// 					 }
	// 				 ],
	// 				 "TxHash": "E102AB0080FDE0DB7A95487ACE99CFDBEF835F366A806A3AFDC8CA6E237033B8"
	// 			 }
	// 			 */
	// 			const method = tx.Method.toLowerCase()
	// 			const commonFields = {
	// 				blockId: baseFileds.height,
	// 				blockTime: baseFileds.time,
	// 				blockHash: baseFileds.hash,

	// 				txHash: tx.TxHash, // -------TxHash
	// 				txid: tx.TxID,

	// 				onchain: 1,
	// 			}
	// 			await that.paserTx(method, commonFields, tx.Detail)
	// 		}
	// 	}

	// 	if (txEvents && txEvents.length > 0) { // 未上链信息 错误的事务信息
	// 		// B80704BB7B4D7C03157D55B5E1B38BEC 查txid 可知道详情
	// 		for (const txEvent of txEvents) {
	// 			// {
	// 			//     "TxID": "EEC34C367674CB741586E63A6DBC5DAC",
	// 			//     "Chaincode": "local",
	// 			//     "Name": "INVOKEERROR",
	// 			//     "Status": 1,
	// 			//     "Detail": "Local invoke error: handling method [MTOKEN.INIT] fail: Can not re-deploy existed data"
	// 			// }

	// 			// 错误的事务 链上查询不到信息  后面保存token等信息表的先不做修改 不应该再考虑onchain =-1 问题
	// 			// {
	// 			//     "jsonrpc": "2.0",
	// 			//     "error": {
	// 			//         "code": 0,
	// 			//         "message": "rpc error: code = Unknown desc = openchain: resource not found",
	// 			//         "data": null
	// 			//     }
	// 			// }

	// 			const tx = {
	// 				blockId: baseFileds.height,
	// 				blockTime: baseFileds.time,
	// 				blockHash: baseFileds.hash,
	// 				method: '',
	// 				txid: txEvent.TxID,
	// 				onchain: -1,
	// 				rawcode: txEvent.Status,
	// 				message: txEvent.Detail,
	// 			}

	// 			const model = that.db.models.Transaction(tx);
	// 			await model.save()
	// 			// const s = await model.save()
	// 			// await that.sendTxMsg(s.toObject())

	// 		}
	// 	}
	// 	await that.sendTxMsg({ txBlockHeight: height })
	// 	// ---------------------------以下是浏览器需要的信息

	// 	const info = await that.db.models.Info.findOne();

	// 	//----------------------------------------------------

	// 	// 更新所读的区块高度
	// 	await that.db.models.Info.findOneAndUpdate({}, {
	// 		$set: {
	// 			height, time: rawdata.time, hash: rawdata.hash,
	// 		},
	// 	}, { upsert: true });

	// 	logger.info(`-------------------第 ${height} 块 end --------------------------`);
	// } else {
	// 	// 查询失败 更新表
	// 	await that.db.models.Info.findOneAndUpdate({}, { $set: { height } }, { upsert: true });
	// 	logger.error(`-------------------第 ${height} 块 读取失败 --------------------------`);
	// 	return
	// }
}

func loopReadAndPaser() {
	dbHeight := initLatestBlockNumber()
	fmt.Println("------------loopReadAndPaser", dbHeight)
	b := make(chan int)
	for {
		select {
		case <-b:
		case <-time.After(time.Second * 1):
			if isNewBlockAvalible(dbHeight) {
				// 解析区块及事务
				readAndParseBlock(dbHeight)
				dbHeight++
			}
		}
		// }
		//  isNewBlockAvalible(dbHeight)
		// 	// 解析区块及事务
		// 	readAndParseBlock(dbHeight)
		// 	dbHeight++
		// }
	}
}
