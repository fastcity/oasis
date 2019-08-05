package controllers

import (
	"century/oasis/api/db"
	"century/oasis/api/models"
	"context"
	"fmt"
	"strings"

	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/astaxie/beego"
	"github.com/astaxie/beego/httplib"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Operations about Users
type AccountController struct {
	beego.Controller
	DB db.MongoInterface
}

func (account *AccountController) Post() {
	defer account.ServeJSON()

	apiKey := account.GetString("apiKey")
	secKey := account.GetString("secKey")
	cbURL := account.GetString("cbUrl")

	op := options.FindOneAndUpdate().SetUpsert(true)

	// 每个用户，每个链
	where := bson.M{"apiKey": apiKey}
	up := bson.M{"$set": bson.M{"apiKey": apiKey, "secKey": secKey, "cbUrl": cbURL}}
	account.DB.ConnCollection("accounts").FindOneAndUpdate(context.Background(), where, up, op)

	account.Data["json"] = map[string]interface{}{
		"code": 0,
		"data": "success",
	}
}

func (account *AccountController) Put() {
	defer account.ServeJSON()

	apiKey := account.GetString("apiKey")
	mode := account.GetString("mode", "setCbUrl")

	where := bson.M{"apiKey": apiKey}

	up := bson.M{}
	switch mode {
	case "setCbUrl":
		cbURL := account.GetString("cbUrl")
		up = bson.M{"$set": bson.M{"cbUrl": cbURL}}
	case "secKey":
		secKey := account.GetString("secKey")
		up = bson.M{"$set": bson.M{"secKey": secKey}}
	}

	// op := options.FindOneAndUpdate().SetUpsert(true)

	account.DB.ConnCollection("accounts").FindOneAndUpdate(context.Background(), where, up)

	account.Data["json"] = map[string]interface{}{
		"code": 0,
		"data": "success",
	}
}

func (account *AccountController) subscribe() {
	defer account.ServeJSON()

	chain := account.GetString("chain")
	addr := account.GetString("addresses")
	if addr == "" {
		account.Data["json"] = map[string]interface{}{
			"code": 0,
			"msg":  "error:address empty",
		}
		return
	}
	op := options.FindOneAndUpdate().SetUpsert(true)
	userID := account.Ctx.Input.GetData("userId")
	_account, _ := primitive.ObjectIDFromHex(userID.(string))
	addrs := strings.Split(addr, ",")

	// 每个用户，每个链
	where := bson.M{"chain": chain, "_account": _account}
	up := bson.M{"$addToSet": bson.M{"addresses": bson.M{"$each": addrs}}, "$set": bson.M{"chain": chain}}
	account.DB.ConnCollection("subscribes").FindOneAndUpdate(context.Background(), where, up, op)

	host := beego.AppConfig.String("builderHost")
	port := beego.AppConfig.String("builderPort")
	url := fmt.Sprintf("http://%s:%s/api/v1/subscribe", host, port)
	reqBody := fmt.Sprintf("addresses=%s&accountId=%s", addr, userID)

	req := httplib.Post(url)
	req.Body(reqBody)
	req.Header("Content-Type", "application/x-www-form-urlencoded")

	// b, _ := req.Bytes()
	var resp models.CommResp

	err := req.ToJSON(&resp)
	if err != nil {
		account.Data["json"] = map[string]interface{}{
			"code": 40000,
			"msg":  err.Error(),
		}

		return
	}
	if resp.Code != 0 {
		account.Data["json"] = map[string]interface{}{
			"code": 40000,
			"msg":  resp.Msg,
		}

		return
	}

	account.Data["json"] = map[string]interface{}{
		"code": 0,
		"data": resp.Data,
	}
}
