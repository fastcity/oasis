package main

import (
	"century/oasis/speed/nodes/eth/jrpc"
	"century/oasis/speed/nodes/eth/models"

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
	chaindb   = "eth"

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
		logger.Error(number, "--isNewBlockAvalible-------------err:", err)
		block = -1
	}
	return block > number

}

func getBlockInfo(number int64, txData interface{}) error {
	return chainConf.GetBlockInfo(number, txData)

	// err := chainConf.GetBlockInfo(number, txData)
	// if err != nil {
	// 	logger.Error("getBlockInfo err", number, err)
	// 	return nil
	// }
	// return b
}

func readAndParseTx(tx models.TXs, number int64) {
	// bytes, err := json.Marshal(tx)

	// if err != nil {
	// 	logger.Error("txid paser error:", err)
	// }
	txid := tx.Txid

	for _, vouts := range tx.Vout {
		v := vouts["value"]

		vountn := vouts["n"]

		script := vouts["scriptPubKey"]

		// scs := script.(map[string]interface{})
		// if addresses, ok := scs["addresses"]; ok {
		// 	addrss := addresses.([]interface{})
		// 	if len(addrss) > 0 {
		// 		address := addrss[0]
		// 		fmt.Println(address)
		// 	}
		// }
		sc, _ := json.Marshal(script)
		address, err := jsonparser.GetString(sc, "addresses", "[0]")
		if err != nil && err != jsonparser.KeyPathNotFoundError {
			logger.Error("jsonparser.GetString vout-->scriptPubKey--->addresses [0] error", err)
		}

		where := bson.D{{"txid", txid}, {"vount", vountn}}
		update := bson.M{"txid": txid, "address": address, "blockHeight": number, "vount": vountn, "value": v, "locked": false, "scriptPubKey": script}

		op := options.FindOneAndUpdate().SetUpsert(true)
		si := db.GetCollection(chaindb, "utxos").FindOneAndUpdate(context.Background(), where, bson.M{"$set": update}, op)
		if si.Err() != nil {
			logger.Error("utxos FindOneAndUpdate error", si.Err())
		}
	}

	for _, vins := range tx.Vin {

		intxid := vins["txid"]
		vout := vins["vout"]
		where := bson.D{{"txid", intxid}, {"vout", vout}}
		db.GetCollection(chaindb, "utxos").FindOneAndDelete(context.Background(), where)
	}

	// jsonparser.ArrayEach(bytes, func(value []byte, dataType jsonparser.ValueType, offset int, err error) {
	// 	v, err := jsonparser.GetFloat(value, "value")
	// 	if err != nil {
	// 		logger.Error("jsonparser.GetString vout-->value error", err)
	// 	}
	// 	vountn, err := jsonparser.GetInt(value, "n")
	// 	if err != nil {
	// 		logger.Error("jsonparser.GetString vout-->scriptPubKey--->addresses [0] error", err)
	// 	}
	// 	address, err := jsonparser.GetString(value, "scriptPubKey", "addresses", "[0]")
	// 	if err != nil {
	// 		logger.Error("jsonparser.GetString vout-->scriptPubKey--->addresses [0] error", err)
	// 	}

	// 	var script map[string]interface{}

	// 	scr, _, _, err := jsonparser.Get(value, "scriptPubKey")
	// 	if err != nil {
	// 		logger.Error("vout get scriptPubKey error:", err)
	// 	}

	// 	err = json.Unmarshal(scr, &script)
	// 	if err != nil {
	// 		logger.Error("scriptPubKey json error:", err)
	// 	}
	// 	where := bson.D{{"txid", txid}, {"vount", vountn}}
	// 	update := bson.M{"txid": txid, "address": address, "blockHeight": number, "vount": vountn, "value": v, "scriptPubKey": script}

	// 	op := options.FindOneAndUpdate().SetUpsert(true)
	// 	si := db.GetCollection(chaindb, "utxos").FindOneAndUpdate(context.Background(), where, bson.M{"$set": update}, op)
	// 	if si.Err() != nil {
	// 		logger.Error("utxos FindOneAndUpdate error", si.Err())
	// 	}

	// }, "vout")
	// logger.Debug("----txid---------over:", txid)

	// jsonparser.ArrayEach(bytes, func(value []byte, dataType jsonparser.ValueType, offset int, err error) {
	// 	intxid, err := jsonparser.GetString(value, "txid")
	// 	if err != nil && err != jsonparser.KeyPathNotFoundError {
	// 		logger.Error("jsonparser.GetString vin------>txid error:", err, dataType.String())
	// 	}
	// 	vout, err := jsonparser.GetInt(value, "vout")
	// 	if err != nil && err != jsonparser.KeyPathNotFoundError {
	// 		logger.Error("jsonparser.GetString vin------>vout error:", err, dataType.String())
	// 	}

	// 	where := bson.D{{"txid", intxid}, {"vout", vout}}
	// 	db.GetCollection(chaindb, "utxos").FindOneAndDelete(context.Background(), where)
	// }, "vin")

	// curor, err := db.GetCollection(chaindb, "utxos").Find(context.Background(), bson.M{"txid": bson.M{"$in": bson.A{intxids}}})

	// if err != nil {
	// 	logger.Error("utxos find ", intxids, "error:", err)
	// }

	// for curor.Next(context.Background()) {

	// }

	// const vin_utxos = await this.db.models.Utxo.find({ txid: { $in: vin_txids } })
	// if (vin_utxos && vin_utxos.length > 0) {
	// 	newVins = vins.map(vin => {
	// 		return vin_utxos.find(p => {
	// 			return p.txid === vin.txid && p.vout == vin.vout
	// 		})
	// 	})
	// 	newVins = newVins.filter(p => (p != null || p != undefined))
	// }

}

func readAndParseBlock(number int64) {
	logger.Debug("--------------------------start paser height:", number)

	blockInfos := models.Blocks{}

	getBlockInfo(number, &blockInfos)

	// 判断回滚?

	// txs := blockInfo.Tx
	// time := blockInfos["time"]

	op := options.FindOneAndUpdate().SetUpsert(true)

	for _, tx := range blockInfos.Tx {
		readAndParseTx(tx, number)

		// for _, v := range tx.Vin {
		// 	vi := models.Vins{}
		// 	json.Unmarshal(v, &vi)
		// 	fmt.Println(vi)
		// }

		transactions := models.Transaction{
			BlockHash:   tx.Hash,
			BlockHeight: number,
			BlockTime:   blockInfos.Time,
			Txid:        tx.Txid,
			Size:        tx.Size,
			Version:     tx.Version,
			Weight:      tx.Weight,
			Vsize:       tx.Vsize,
			Vins:        tx.Vin,
			Vouts:       tx.Vout,
		}

		where := bson.M{"txid": tx.Txid}

		db.GetCollection(chaindb, "transactions").FindOneAndUpdate(context.Background(), where, bson.M{"$set": transactions}, op)
	}

	// if kafka != nil {
	// 	kafka.SendMsg("TX", strings.ToUpper(chaindb)+"_TX", number)
	// } else {
	// 	logger.Error("kafka error , can not send")
	// }

	db.GetCollection(chaindb, "infos").FindOneAndUpdate(context.Background(), bson.M{}, bson.M{"$set": bson.M{"height": number}}, op)

	logger.Debug(number, "--------------------------end-------------")

}

func loopReadAndPaser() {

	timer := time.NewTicker(time.Millisecond * 1000)
	dbHeight := initLatestBlockNumber() //dbHeight 弄一个channel
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
