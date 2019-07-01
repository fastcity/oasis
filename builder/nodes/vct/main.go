package main

import (
	"century/oasis/builder/nodes/vct/util/chain"
	"century/oasis/builder/nodes/vct/util/comm"
	"century/oasis/builder/nodes/vct/util/dbs"
	"century/oasis/builder/nodes/vct/util/dbs/models"
	"context"
	"flag"
	"fmt"
	"math/big"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/buger/jsonparser"
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
	kModel = comm.NewConsumer(viper.GetStringSlice("kafka.service"))
	// if kModel.Consumer == nil {
	// 	fmt.Println("init kafka fail-------")
	// }
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

func loopReadAndPaser() {
	// dbHeight := initLatestBlockNumber()
	// b := make(chan int)
	// for {
	// 	select {
	// 	case <-b:
	// 	case <-time.After(time.Second * 1):
	// 		// if isNewBlockAvalible(dbHeight) {
	// 		// 	// 解析区块及事务
	// 		// 	readAndParseBlock(dbHeight)
	// 		// 	dbHeight++
	// 		// }
	// 		kModel.ReciveMsg()
	// 	}
	// }
	msg := make(chan []byte)
	go func() {
		kModel.ReciveMsg(msg)
	}()

	for {
		select {

		case m := <-msg:
			paserTx(m)
			// default:
			// 	fmt.Println("++++++++++")
		}
	}

	// kModel.ReciveMsg()
}

func paserTx(msg []byte) {
	fmt.Println("-+++++++++", string(msg))

	h := string(msg)
	// fmt.Println("-+++++++++", string(blockInfo))
	type res struct {
		Result *models.Blocks `json:"result"`
	}
	// fmt.Println("--------------- blockInfo", string(blockInfos))
	coror, err := db.GetCollection("vct", "transactions").Find(context.Background(), bson.M{"blockheight": h})

	if err != nil {
		fmt.Println("get transferaction error")
	}

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
			txs.Txid = item.Txid
			txs.Log = item.Detail
			txs.OnChain = false

		}

		op := options.FindOneAndUpdate().SetUpsert(true)

		// docs := bson.M{
		// 	"height": b.Result.Height,
		// 	"hash":   b.Result.Hash,
		// 	"time":   b.Result.TimeStamp,
		// }
		// _, err = db.GetCollection("vct", "blocks").FindOneAndUpdate(context.Background(), bson.M{}, bson.M{"$set": docs}, op)
		// if err != nil {
		// 	fmt.Println("insert one err", err)
		// }

		db.GetCollection("vct", "transactions").FindOneAndUpdate(context.Background(), bson.M{"txid": b.Result.Txid}, bson.M{"$set": txs}, op)
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
