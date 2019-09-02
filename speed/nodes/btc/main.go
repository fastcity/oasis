package main

import (
	"century/oasis/speed/nodes/btc/jrpc"
	"century/oasis/speed/nodes/btc/models"

	"century/oasis/speed/util"
	"context"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/buger/jsonparser"
	jsoniter "github.com/json-iterator/go"
	"github.com/spf13/viper"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.uber.org/zap"
	// "century/oasis/speed/vct/db"
)

var (
	chain     string
	env       string
	json      = jsoniter.ConfigCompatibleWithStandardLibrary
	db        util.MongoI
	chainConf jrpc.ChainApi
	kafka     util.KaInterface
	assign    = "TOKEN.ASSIGN"
	chaindb   = "btc"

	logger *zap.SugaredLogger
)

// Result 返回结果
type Result struct {
	Code int                    `json:code`
	Data map[string]interface{} `json:Data`
}

func main() {

	flag.StringVar(&chain, "chain", chaindb, "chain")
	flag.StringVar(&env, "env", "dev", "env")
	flag.Parse()

	logger = util.NewLogger()
	beforeStart()
	defer kafka.Close()
	defer db.Close()
	loopReadAndPaser()

}

func beforeStart() {

	initConf()
	initKafka()
	initDB()
	initNodeAPI()

}

func initConf() {
	// confirmedNumber = viper.GetInt("chain.confirmedNumber")
	defaultConf := filepath.Join("../../config/", strings.ToLower(env), "nodes")
	gopath := os.Getenv("GOPATH")
	pathConf := []string{defaultConf, "."}
	for _, p := range filepath.SplitList(gopath) {
		path := filepath.Join(p, "src/century/oasis/builder/config", strings.ToLower(env), "nodes")
		pathConf = append(pathConf, path)
	}

	InitViper(strings.ToLower(chain), strings.ToLower(chain), pathConf)

}

func initDB() {
	db = util.NewDBs(viper.GetString("db.addr"))
	err := db.GetConn()
	if err != nil {
		fmt.Println("connect mongo error", err)
	}
	chaindb = viper.GetString("chain.chaindb")
}

func initKafka() {

	kafka = util.NewProducer(viper.GetStringSlice("kafka.service"))

}

func initNodeAPI() {
	api := fmt.Sprintf("%s://%s:%s", viper.GetString("node.protocal"), viper.GetString("node.host"), viper.GetString("node.port"))
	username := viper.GetString("node.auth.user")
	pwd := viper.GetString("node.auth.pass")
	chainConf = jrpc.NewChainAPi(api, username, pwd)
}

//InitViper we can set viper which fabric peer is used
func InitViper(envprefix string, filename string, configPath []string) error {
	logger.Info("envprefix:", envprefix, ",filename:", filename, ",configPath:", configPath)

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

	var result models.Info

	ctx := context.Background()

	err := db.GetCollection(chaindb, "infos").FindOne(ctx, bson.M{}).Decode(&result)
	if err != nil {
		logger.Error(" collection.FindOne err", err)
	}

	logger.Debug("initLatestBlockNumber", result)
	if result.Height > 0 {
		return result.Height + 1
	}
	return 0
}

func isNewBlockAvalible(number int64) bool {

	block, err := chainConf.GetBlockHeight()
	if err != nil {
		logger.Error(number, "isNewBlockAvalible-------------err", err)
		block = -1
	}
	return block > number

}

func getBlockInfo(number int64) map[string]interface{} {

	b, err := chainConf.GetBlockInfo(number)
	if err != nil {
		logger.Error("getBlockInfo err", number, err)
		return nil
	}
	return b
}

func readAndParseTx(tx interface{}, number int64) {
	// const { txid, vout } = transaction
	// const voutCount = vout.length

	// // 没有设置离线utxo 或者 当前块高>离线utxo block时需要解析出新的utxo
	// const voutUtxos = []
	// if ((this.offline_utxo_block < 0) || ((this.offline_utxo_block >= 0) && (number >= this.offline_utxo_block))) {
	// 	for (const v of vout) {
	// 		const { value, n, scriptPubKey } = v
	// 		if (!scriptPubKey.addresses || scriptPubKey.addresses.length < 1) continue
	// 		const address = scriptPubKey.addresses[0]

	// 		// await this.db.models.Utxo.findOneAndUpdate({ txid, value, vout: n, voutCount, address, blockHeight: number }, { $set: { txid, value, vout: n, voutCount, address, scriptPubKey, blockHeight: number, locked: false } }, { upsert: true })
	// 		voutUtxos.push({ updateOne: { 'filter': { txid, value, vout: n, voutCount, address, blockHeight: number }, 'update': { $set: { txid, value, vout: n, voutCount, address, scriptPubKey, blockHeight: number, locked: false } }, 'upsert': true } })
	// 	}

	// 	if (voutUtxos.length > 0) {
	// 		await this.db.models.Utxo.bulkWrite(voutUtxos)
	// 	}
	// }
	bytes, _ := json.Marshal(tx)
	txid, err := jsonparser.GetString(bytes, "txid")

	if err != nil {
		logger.Error("txid paser error:", err)
	}

	jsonparser.ArrayEach(bytes, func(value []byte, dataType jsonparser.ValueType, offset int, err error) {
		v, err := jsonparser.GetFloat(value, "value")
		if err != nil {
			logger.Error("jsonparser.GetString vout-->value error", err)
		}
		vountn, err := jsonparser.GetInt(value, "n")
		if err != nil {
			logger.Error("jsonparser.GetString vout-->scriptPubKey--->addresses [0] error", err)
		}
		address, err := jsonparser.GetString(value, "scriptPubKey", "addresses", "[0]")
		if err != nil {
			logger.Error("jsonparser.GetString vout-->scriptPubKey--->addresses [0] error", err)
		}

		var script map[string]interface{}

		scr, _, _, err := jsonparser.Get(value, "scriptPubKey")
		if err != nil {
			logger.Error("vout get scriptPubKey error:", err)
		}

		err = json.Unmarshal(scr, &script)
		if err != nil {
			logger.Error("scriptPubKey json error:", err)
		}
		where := bson.D{{"txid", txid}, {"address", address}, {"blockHeight", number}}
		update := bson.M{"txid": txid, "address": address, "blockHeight": number, "vount": vountn, "value": v, "scriptPubKey": script}

		op := options.FindOneAndUpdate().SetUpsert(true)
		si := db.GetCollection(chaindb, "utxos").FindOneAndUpdate(context.Background(), where, bson.M{"$set": update}, op)
		if si.Err() != nil {
			logger.Error("utxos FindOneAndUpdate error", si.Err())
		}

	}, "vout")
	logger.Debug("----txid---------over:", txid)
}

func readAndParseBlock(number int64) {
	blockInfos := getBlockInfo(number)

	txs := blockInfos["tx"]

	switch ts := txs.(type) {
	case []interface{}:
		for _, tx := range ts {
			readAndParseTx(tx, number)
		}

	}

	// h, err := jsonparser.GetString(blockInfos, "result", "Height")
	// if err != nil {
	// 	logger.Error("jsonparser.GetString error", err)
	// }
	// he, _ := strconv.Atoi(h)
	// logger.Debug("-----jsonparser. Height", he)

	// if b.Result.Height != "" {

	// 	txs := models.Transaction{
	// 		BlockHeight: b.Result.Height,
	// 		BlockTime:   b.Result.TimeStamp,
	// 		BlockHash:   b.Result.Hash,
	// 		OnChain:     true,
	// 	}

	// 	for _, item := range b.Result.Txs {
	// 		txs.Txid = item.Txid
	// 		txs.Method = item.Method
	// 		if item.Method == "batch" {
	// 			jsonparser.ArrayEach(blockInfos, func(value []byte, dataType jsonparser.ValueType, offset int, err error) {
	// 				m, err := jsonparser.GetString(value, "Method")
	// 				if err != nil {
	// 					logger.Error("jsonparser.GetString Method error", err)
	// 				}
	// 				if m == assign {
	// 					to, _ := jsonparser.GetString(value, "Method", "to")
	// 					token, _ := jsonparser.GetString(value, "Method", "name")
	// 					txs.To = to
	// 					txs.TokenKey = token
	// 				}

	// 			}, "result", "Transactions", "Detail")

	// 		} else {
	// 			logger.Debug("-------------item.Detail  not batch", item.Detail)
	// 			txs.From = item.Detail.From
	// 			txs.To = item.Detail.To
	// 			txs.Value = item.Detail.Amount.String()
	// 			txs.TokenKey = item.Detail.Token
	// 		}
	// 	}

	// 	for _, item := range b.Result.Events {
	// 		// txs.From=item.
	// 		txs.Txid = item.Txid
	// 		txs.Log = item.Detail
	// 		txs.OnChain = false

	// 	}

	// 	op := options.FindOneAndUpdate().SetUpsert(true)
	// 	// rs := db.GetCollection(chaindb, "infos").FindOneAndUpdate(context.Background(), bson.M{}, bson.D{{"$set", bson.M{"height": h}}}, op)
	// 	rs := db.GetCollection(chaindb, "infos").FindOneAndUpdate(context.Background(), bson.M{}, bson.M{"$set": bson.M{"height": he}}, op)
	// 	if rs.Err() != nil {
	// 		logger.Error("FindOneAndUpdate err", rs.Err())
	// 	}

	// 	db.GetCollection(chaindb, "transactions").FindOneAndUpdate(context.Background(), bson.M{"txid": b.Result.Txid}, bson.M{"$set": txs}, op)

	// 	topic := strings.ToUpper(chaindb)
	// 	kafka.SendMsg("TX", topic+"_TX", h)
	// }
}

func loopReadAndPaser() {

	timer := time.NewTicker(time.Millisecond * 1000)
	dbHeight := initLatestBlockNumber()
	for {
		select {

		case <-time.After(time.Second * 3):
			// 3s 同步一次数据库的值
			dbHeight = initLatestBlockNumber()

		case <-timer.C:
			if isNewBlockAvalible(dbHeight) {
				// 解析区块及事务
				readAndParseBlock(dbHeight)
				dbHeight++
			}
		}
	}
}
