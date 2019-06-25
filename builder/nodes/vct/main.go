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

	"github.com/Shopify/sarama"
	"github.com/buger/jsonparser"
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
	// host := viper.GetString("service.host")
	// port := viper.GetString("service.port")
	// router(host + ":" + port)

	db = dbs.New("127.0.01", 27017)
	err := db.GetConn()
	if err != nil {
		fmt.Println("connect mongo error", err)
	}
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
	u := util.Util{BaseURL: "http://127.0.0.1:7080/api/v1"}

	block, err := u.GetBlockHeight()
	if err != nil {
		fmt.Println("--------------err", err)
		block = -1
	}
	return block > number

}

func getBlockInfo(number int64) []byte {
	u := util.Util{BaseURL: "http://127.0.0.1:7080/api/v1"}

	b, err := u.GetBlockInfo(number)

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

	coll := db.GetCollection("vct", "transactions")
	err = coll.Drop(context.Background())

	if b.Result.Height != "" {

		// result, err := coll.InsertOne(
		// 	context.Background(),
		// 	bson.D{
		// 		{"item", "canvas"},
		// 		{"qty", 100},
		// 		{"tags", bson.A{"cotton"}},
		// 		{"size", bson.D{
		// 			{"h", 28},
		// 			{"w", 35.5},
		// 			{"uom", "cm"},
		// 		}},
		// 	})
		docs := bson.D{
			{"height", b.Result.Height},
			{"hash", b.Result.Hash},
			{"time", b.Result.TimeStamp},
		}
		result, err := coll.InsertOne(context.Background(), docs)
		if err != nil {

		}
		fmt.Println("insert one ", result.InsertedID)
	}

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

func kakfa() {
	config := sarama.NewConfig()
	config.Producer.RequiredAcks = sarama.WaitForAll          //赋值为-1：这意味着producer在follower副本确认接收到数据后才算一次发送完成。
	config.Producer.Partitioner = sarama.NewRandomPartitioner //写到随机分区中，默认设置8个分区
	config.Producer.Return.Successes = true
	msg := &sarama.ProducerMessage{}
	msg.Topic = `nginx_log`
	msg.Value = sarama.StringEncoder("this is a good test")
	client, err := sarama.NewSyncProducer([]string{"127.0.0.1:9092"}, config)
	if err != nil {
		fmt.Println("producer close err, ", err)
		return
	}
	defer client.Close()
	pid, offset, err := client.SendMessage(msg)

	if err != nil {
		fmt.Println("send message failed, ", err)
		return
	}
	fmt.Printf("分区ID:%v, offset:%v \n", pid, offset)

}
