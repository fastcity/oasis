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
}

func getBlockInfoByHash(hash string, info interface{}) error {
	return chainConf.GetTransactionReceipt(hash, info)
}

func parseTx(txHex *models.TransactionHex, number int64) {

	var txHexM models.TransactionHex

	logger.Debug("parseTx----> getBlockInfoByHash txid:", txHex.Txid)
	err := getBlockInfoByHash(txHex.Txid, &txHexM)
	if err != nil {
		logger.Error("parseTx----> getBlockInfoByHash error:", err)
	}
	txHex.TokenKey = txHexM.TokenKey
	txHex.GasUsed = txHexM.GasUsed
	txHex.Status = txHexM.Status
	txHex.Logs = txHexM.Logs

	raw := txHex.HexToRaw()

	op := options.FindOneAndUpdate().SetUpsert(true)
	where := bson.M{"txid": txHex.Txid}
	db.GetCollection(chaindb, "transactions").FindOneAndUpdate(context.Background(), where, bson.M{"$set": raw}, op)

	logger.Debug("parseTx----> getBlockInfoByHash get txHex", txHex)
}

func readAndParseBlock(number int64) {
	logger.Debug("--------------------------start paser height:", number)

	blockInfosHex := models.BlocksHex{}

	getBlockInfo(number, &blockInfosHex)

	// blockInfos := blockInfosHex.HexToRaw()
	// 判断回滚?

	op := options.FindOneAndUpdate().SetUpsert(true)

	for _, tx := range blockInfosHex.Transactions {
		tx.BlockTime = blockInfosHex.Timestamp
		parseTx(&tx, number)

	}
	// for _, tx := range blockInfos.Tx {
	// 	readAndParseTx(tx, number)

	// 	transactions := models.Transaction{
	// 		BlockHash:   tx.Hash,
	// 		BlockHeight: number,
	// 		BlockTime:   blockInfos.Time,
	// 		Txid:        tx.Txid,
	// 		Size:        tx.Size,
	// 		Version:     tx.Version,
	// 		Weight:      tx.Weight,
	// 		Vsize:       tx.Vsize,
	// 		Vins:        tx.Vin,
	// 		Vouts:       tx.Vout,
	// 	}

	// 	where := bson.M{"txid": tx.Txid}

	// 	db.GetCollection(chaindb, "transactions").FindOneAndUpdate(context.Background(), where, bson.M{"$set": transactions}, op)
	// }

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
