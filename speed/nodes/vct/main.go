package main

import (
	"century/oasis/speed/nodes/vct/models"
	gchain "century/oasis/speed/nodes/vct/util/chain"
	"century/oasis/speed/util"
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
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.uber.org/zap"
	// "century/oasis/speed/vct/db"
)

var (
	chain     string
	env       string
	json      = jsoniter.ConfigCompatibleWithStandardLibrary
	db        util.MongoI
	chainConf gchain.ChainApi
	kafka     util.KaInterface
	assign    = "TOKEN.ASSIGN"
	chaindb   = "vct"

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
	// kafka = util.NewConsumer(viper.GetStringSlice("kafka.service"))
	// kafka.SetTopics(viper.GetStringSlice("kafka.topics"))
	// kafka.SetKeys(viper.GetStringSlice("kafka.keys"))

}

func initNodeAPI() {
	api := fmt.Sprintf("%s://%s:%s", viper.GetString("node.protocal"), viper.GetString("node.host"), viper.GetString("node.port"))
	chainConf = gchain.NewChainAPi(api)

}

func getBlockHeight(w http.ResponseWriter, r *http.Request) {

	h, err := chainConf.GetBlockHeight()
	if err != nil {
		// fmt.Println("--------------err", err)
		logger.Error("--------------err", err)
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
		// log.Fatal(err)
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
		logger.Error("--------------err", err)
		block = -1
	}
	return block > number

}

func getBlockInfo(number int64) []byte {

	b, err := chainConf.GetBlockInfo(number)
	if err != nil {
		logger.Error("getBlockInfo err", number, err)
		return nil
	}
	return b
}

func readAndParseBlock(number int64) {
	blockInfos := getBlockInfo(number)
	// fmt.Println("-+++++++++", string(blockInfo))
	type res struct {
		Result *models.Blocks `json:"result"`
	}
	// fmt.Println("--------------- blockInfo", string(blockInfos))
	b := &res{}
	err := json.Unmarshal(blockInfos, b)
	if err != nil {
		// return -1, err
		logger.Error("!!!!!!! --------json.Unmarshal blockInfo error", err)
	}
	// fmt.Println("---------------json.Unmarshal ", b)

	h, err := jsonparser.GetString(blockInfos, "result", "Height")
	if err != nil {
		logger.Error("jsonparser.GetString error", err)
	}
	he, _ := strconv.Atoi(h)
	logger.Debug("-----jsonparser. Height", he)

	if b.Result.Height != "" {

		txs := models.Transaction{
			BlockHeight: b.Result.Height,
			BlockTime:   b.Result.TimeStamp,
			BlockHash:   b.Result.Hash,
			OnChain:     true,
		}

		for _, item := range b.Result.Txs {
			txs.Txid = item.Txid
			txs.Method = item.Method
			if item.Method == "batch" {
				jsonparser.ArrayEach(blockInfos, func(value []byte, dataType jsonparser.ValueType, offset int, err error) {
					m, err := jsonparser.GetString(value, "Method")
					if err != nil {
						logger.Error("jsonparser.GetString Method error", err)
					}
					if m == assign {
						to, _ := jsonparser.GetString(value, "Method", "to")
						token, _ := jsonparser.GetString(value, "Method", "name")
						txs.To = to
						txs.TokenKey = token
					}

				}, "result", "Transactions", "Detail")

			} else {
				logger.Debug("-------------item.Detail  not batch", item.Detail)
				txs.From = item.Detail.From
				txs.To = item.Detail.To
				txs.Value = item.Detail.Amount.String()
				txs.TokenKey = item.Detail.Token
			}
		}

		for _, item := range b.Result.Events {
			// txs.From=item.
			txs.Txid = item.Txid
			txs.Log = item.Detail
			txs.OnChain = false

		}

		op := options.FindOneAndUpdate().SetUpsert(true)
		// rs := db.GetCollection(chaindb, "infos").FindOneAndUpdate(context.Background(), bson.M{}, bson.D{{"$set", bson.M{"height": h}}}, op)
		rs := db.GetCollection(chaindb, "infos").FindOneAndUpdate(context.Background(), bson.M{}, bson.M{"$set": bson.M{"height": he}}, op)
		if rs.Err() != nil {
			logger.Error("FindOneAndUpdate err", rs.Err())
		}

		db.GetCollection(chaindb, "transactions").FindOneAndUpdate(context.Background(), bson.M{"txid": b.Result.Txid}, bson.M{"$set": txs}, op)

		kafka.SendMsg("TX", "VCT_TX", h)
	}

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

	// 			// {
	// 			//     "TxID": "EEC34C367674CB741586E63A6DBC5DAC",
	// 			//     "Chaincode": "local",
	// 			//     "Name": "INVOKEERROR",
	// 			//     "Status": 1,
	// 			//     "Detail": "Local invoke error: handling method [MTOKEN.INIT] fail: Can not re-deploy existed data"
	// 			// }

}

func loopReadAndPaser() {
	dbHeight := initLatestBlockNumber()
	for {
		select {

		case <-time.After(time.Second * 1):
			if isNewBlockAvalible(dbHeight) {
				// 解析区块及事务
				readAndParseBlock(dbHeight)
				dbHeight++
			}
		}
	}
}
