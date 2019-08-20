package controllers

import (
	"century/oasis/api/db"
	"century/oasis/api/models"
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/astaxie/beego"
	"github.com/astaxie/beego/httplib"
	"github.com/buger/jsonparser"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Operations about Users
type TransferController struct {
	beego.Controller
	DB db.MongoInterface
}

//CreateTransferTxData 创建未签名事务
func (tf *TransferController) CreateTransferTxData() {
	defer tf.ServeJSON()
	transfer := models.Transfer{}
	// json.Unmarshal(tf.Ctx.Input.RequestBody, &transfer) // TODO: 不好使,请求头需为json
	transfer.From = tf.GetString("from")
	transfer.To = tf.GetString("to")
	transfer.Value = tf.GetString("value")
	transfer.TokenKey = tf.GetString("tokenKey", "-")
	transfer.Chain = tf.GetString("chain")
	transfer.Coin = tf.GetString("coin")
	transfer.CreateID = tf.GetString("createId")

	// tf.Ctx.Input.RequestBody

	insertresult, err := tf.DB.ConnCollection("transfers").InsertOne(context.Background(),
		bson.M{"chain": transfer.Chain, "coin": transfer.Coin, "requestBody": transfer, "createId": transfer.CreateID})
	if err != nil {
		tf.Data["json"] = map[string]interface{}{
			"code": 40001,
			"msg":  err.Error(),
		}

		return
	}

	switch str := insertresult.InsertedID.(type) {
	case primitive.ObjectID:
		transfer.RequestID = str.Hex()
	}

	// transfer.RequestID = insertresult.InsertedID
	// const transferM = new ctx.model.Transfer({
	// 	chain, coin, tokenKey, _account: ctx.curUser._id, requestBody: body, createId,
	// })
	// const transferSaved = await transferM.save()
	// reqBody := fmt.Sprintf("from=%s&to=%s&value=%s", transfer.From, transfer.To, transfer.Value)

	// resp, err := http.Post(url, "application/x-www-form-urlencoded", strings.NewReader(reqBody))

	// if err != nil {
	// 	fmt.Println("get err", err)
	// }
	// defer resp.Body.Close()

	// result, _ := ioutil.ReadAll(resp.Body)
	// fmt.Println(string(result))
	host := beego.AppConfig.String("builderHost")
	port := beego.AppConfig.String("builderPort")
	url := fmt.Sprintf("http://%s:%s/api/v1/createTransferTxData", host, port)

	req := httplib.Post(url)
	// req.Header("Content-Type", "application/json")
	req.JSONBody(transfer)

	var resp models.CommResp
	reqb, err := req.Bytes()
	v, err := jsonparser.Set(reqb, []byte(`"`+transfer.RequestID+`"`), "data", "requestId")
	if err != nil {
		tf.Data["json"] = map[string]interface{}{
			"code": 40002,
			"msg":  err.Error(),
		}

		return
	}
	// fmt.Println("v", string(v), len(reqb))
	// fmt.Println("v", string(reqb), len(reqb))
	// // json
	// re, _ := req.Response()
	// fmt.Println("v", string(reqb), re.ContentLength)
	// err = req.ToJSON(&resp) // 被截断，content-lenght??
	err = json.Unmarshal(v, &resp)

	// fmt.Println(resp.Data)
	if err != nil {
		tf.Data["json"] = map[string]interface{}{
			"code": 40001,
			"data": resp.Data,
			"msg":  err.Error(),
		}

		return
	}
	if resp.Code != 0 {
		tf.Data["json"] = map[string]interface{}{
			"code": 40000,
			"data": resp.Data,
			"msg":  resp.Msg,
		}

		return
	}
	txData := resp.Data["txData"]
	// id, _ := primitive.ObjectIDFromHex(transfer.RequestID)
	// where := bson.M{"_id": id} //insertresult.InsertedID
	where := bson.M{"_id": insertresult.InsertedID} //
	updateStr := bson.M{"$set": bson.M{"txData": txData, "code": 2, "status": "txDataCreated", "updatedAt": time.Now().Unix()}, "$push": bson.M{"logs": "update txData at: ${Date.now()}"}}
	upresult := tf.DB.ConnCollection("transfers").FindOneAndUpdate(context.Background(), where, updateStr)
	if upresult.Err() != nil {
		tf.Data["json"] = map[string]interface{}{
			"code": 40001,
			"msg":  upresult.Err().Error(),
		}

		return
	}

	tf.Data["json"] = map[string]interface{}{
		"code": 0,
		"data": resp.Data,
	}
}

//SubmitTx 提交签名事务
func (tf *TransferController) SubmitTx() {
	defer tf.ServeJSON()

	// d := tf.Data["db"].(db.MongoInterface)

	// var transfer models.Transfer
	// json.Unmarshal(tf.Ctx.Input.RequestBody, &transfer) // 不好使 传 json TODO: 带研究
	singedRawTx := tf.GetString("singedRawTx")
	requestID := tf.GetString("requestId")

	// reqBody := fmt.Sprintf("from=%s&to=%s&value=%s", transfer.From, transfer.To, transfer.Value)

	// resp, err := http.Post(url, "application/x-www-form-urlencoded", strings.NewReader(reqBody))

	// if err != nil {
	// 	fmt.Println("get err", err)
	// }
	// defer resp.Body.Close()

	// result, _ := ioutil.ReadAll(resp.Body)
	// fmt.Println(string(result))
	host := beego.AppConfig.String("builderHost")
	port := beego.AppConfig.String("builderPort")
	url := fmt.Sprintf("http://%s:%s/api/v1/submitTxData", host, port)
	reqBody := fmt.Sprintf("requestId=%s&singedRawTx=%s", requestID, singedRawTx)
	req := httplib.Post(url)
	req.Body(reqBody)
	req.Header("Content-Type", "application/x-www-form-urlencoded")

	var resp models.CommResp
	err := req.ToJSON(&resp)
	if err != nil {
		tf.Data["json"] = map[string]interface{}{
			"code": 40001,
			"msg":  err.Error(),
		}

		return
	}
	if resp.Code != 0 {
		tf.Data["json"] = map[string]interface{}{
			"code": 40000,
			"msg":  resp.Msg,
		}

		return
	}
	// TODO: kakfa
	//  const updateStr2 = { $set: { kafkaMsg, kafkaMsgResult, code: 8, status: 'sendKafkaMsg', updatedAt: Date.now() }, $push: { logs: `sendKafkaMsg at: ${Date.now()}` } }
	updateStr := bson.M{"$set": bson.M{"code": 8, "status": "send", "updatedAt": time.Now().Unix()}, "$push": bson.M{"logs": "sendKafkaMsg at: ${Date.now()}"}}
	upresult := tf.DB.ConnCollection("transfers").FindOneAndUpdate(context.Background(), bson.M{"_id": requestID}, updateStr)
	if upresult.Err() != nil {
		tf.Data["json"] = map[string]interface{}{
			"code": 40001,
			"msg":  upresult.Err().Error(),
		}

		return
	}
	tf.Data["json"] = map[string]interface{}{
		"code": 0,
		"data": struct {
			RequestID string `json:"requestId"`
		}{requestID},
		"msg": "create task success, the status of task will be notify",
	}
}

func (tf *TransferController) GetTxStatus() {
	defer tf.ServeJSON()
	requestID := tf.GetString("requestId")

	id, err := primitive.ObjectIDFromHex(requestID)
	if err != nil {
		tf.Data["json"] = map[string]interface{}{
			"code": 40001,
			"msg":  err.Error(),
		}
	}

	result := tf.DB.ConnCollection("transfers").FindOne(context.Background(), bson.M{"_id": id})
	if result.Err() != nil {
		tf.Data["json"] = map[string]interface{}{
			"code": 40001,
			"msg":  result.Err().Error(),
		}

		return
	}

	var res map[string]interface{}
	result.Decode(&res)

	tf.Data["json"] = map[string]interface{}{
		"code": 0,
		"data": res,
	}
}
