package main

import (
	"century/oasis/speed/nodes/vct/dbs"
	"century/oasis/speed/nodes/vct/dbs/models"
	"century/oasis/speed/nodes/vct/util"
	"context"
	"flag"
	"fmt"
	"math/big"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	// "github.com/Shopify/sarama"
	"github.com/json-iterator/go"
	"github.com/spf13/viper"
	"go.mongodb.org/mongo-driver/bson"
	// "century/oasis/speed/vct/db"
)

var (
	chain              string
	env                string
	json               = jsoniter.ConfigCompatibleWithStandardLibrary
	db                 *dbs.Conn
	currentBlockNumber = 0
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
		path := filepath.Join(p, "src/century/oasis/config", strings.ToLower(env), "nodes")
		// viper.AddConfigPath(path)
		InitViper(strings.ToLower(chain), strings.ToLower(chain), path)
	}
	host := viper.GetString("service.host")
	port := viper.GetString("service.port")
	router(host + ":" + port)

	db = dbs.New("127.0.01", 27017)
	err := db.GetConn()
	if err != nil {
		fmt.Println("connect mongo error", err)
	}
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
	u := util.Util{BaseURL: "127.0.0.1:7080"}
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
	u.CreateTransactionData(from, to, tokenKey, amount)

}

func submitTxDtaHandler(w http.ResponseWriter, r *http.Request) {
	u := util.Util{BaseURL: "127.0.0.1:7080"}
	from := r.PostFormValue("from")
	to := r.PostFormValue("to")

	tokenKey := r.PostFormValue("tokenKey")

	amount, ok := big.NewInt(0).SetString(r.PostFormValue("value"), 0)

	if !ok || (amount.IsUint64() && amount.Uint64() == 0) {
		// s.NormalErrorF(rw, 0, "Invalid amount")
		fmt.Println("Invalid amount")
		return
	}
	u.CreateTransactionData(from, to, tokenKey, amount)

}

func getBlockHeight(w http.ResponseWriter, r *http.Request) {
	u := util.Util{BaseURL: "http://127.0.0.1:7080/api/v1"}

	h, err := u.GetBlockHeight()
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

func initLatestBlockNumber() int {

	collection := db.GetCollection("Info")

	result := &models.Info{}

	filter := bson.M{}
	ctx, _ = context.WithTimeout(context.Background(), 5*time.Second)
	err = collection.FindOne(ctx, filter).Decode(&result)
	if err != nil {
		// log.Fatal(err)
	}

	if info.Height > 0 {
		return info.Height + 1
	}
	return 0
	// 	const info = this.db.models.Info.findOne()
	// 	this.latestBlockNumber = info ? info.height + 1 : 0
}

func isNewBlockAvalible(number int64) {

	const block = getBlockCount()
	if block.result {
		const top = block.result.Height // 最新的高度
		logger.debug(`get astro 最新区块高度:${top}； 开始查询： ${number}`)
		return top > number

	}
	logger.error(`get astro 最新区块高度 出错。 当前打算查询块： ${number}`)
	return false
}

func getBlockCount() {
	u := util.Util{BaseURL: "http://127.0.0.1:7080/api/v1"}

	h, err := u.GetBlockHeight()
	if err != nil {
		fmt.Println("--------------err", err)
		return -1
	}

	return h
}
