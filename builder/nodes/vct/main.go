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
	http.HandleFunc("/api/v1/createTransferTxDta", createTransactionDataHandler)
	http.HandleFunc("/api/v1/submitTxDta", submitTxDtaHandler)
	http.HandleFunc("/api/v1/getBlockHeight", getBlockHeight)

	err := http.ListenAndServe(":7799", nil)
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
