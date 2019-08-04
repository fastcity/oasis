package controllers

import (
	"century/oasis/api/db"
	"century/oasis/api/models"
	"context"
	"fmt"
	"strings"

	"github.com/astaxie/beego"
	"github.com/astaxie/beego/httplib"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Operations about Users
type AccountController struct {
	beego.Controller
	DB db.MongoInterface
}

func (account *AccountController) subscribe() {
	defer account.ServeJSON()
	addr := account.GetString("addresses")
	if addr == "" {
		account.Data["json"] = map[string]interface{}{
			"code": 0,
			"msg":  "error:address empty",
		}
		return
	}
	addrs := strings.Split(addr, ",")
	for _, add := range addrs {
		id, _ := primitive.ObjectIDFromHex(add) //TODO: user id
		where := bson.M{"_id": id}
		up := bson.M{"$addToSet": bson.M{"addresses": add}}
		account.DB.ConnCollection("subscribes").FindOneAndUpdate(context.Background(), where, up)
	}

	host := beego.AppConfig.String("builderHost")
	port := beego.AppConfig.String("builderPort")
	url := fmt.Sprintf("http://%s:%s/api/v1/subscribe", host, port)
	reqBody := fmt.Sprintf("addresses=%s&accountId=%s", addr, addr) //TODO: user id

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
