package controllers

import (
	"century/oasis/api/models"
	"fmt"

	"github.com/astaxie/beego"
	"github.com/astaxie/beego/httplib"
)

// Operations about Users
type BalanceController struct {
	beego.Controller
}

func (balance *BalanceController) Get() {
	// var balance models.Balance
	// json.Unmarshal(u.Ctx.Input.RequestBody, &balance)
	defer balance.ServeJSON()
	addr := balance.GetString("address")
	if addr == "" {
		balance.Data["json"] = map[string]interface{}{
			"code": 0,
			"msg":  "error:address empty",
		}
		return
	}

	host := beego.AppConfig.String("builderHost")
	port := beego.AppConfig.String("builderPort")
	url := fmt.Sprintf("http://%s:%s/api/v1/getBalance?address=%s", host, port, addr)

	req := httplib.Get(url)
	b, _ := req.Bytes()
	fmt.Println("url", url, "ba", b)
	var resp models.CommResp

	err := req.ToJSON(&resp)
	if err != nil {
		balance.Data["json"] = map[string]interface{}{
			"code": 40000,
			"msg":  err.Error(),
		}

		return
	}
	if resp.Code != 0 {
		balance.Data["json"] = map[string]interface{}{
			"code": 40000,
			"msg":  resp.Msg,
		}

		return
	}

	balance.Data["json"] = map[string]interface{}{
		"code": 0,
		"data": resp.Data,
	}

	// resp, err := http.Get(url)
	// if err != nil {

	// }
	// resp.Body
	// uid := models.AddUser(user)
	// u.Data["json"] = map[string]string{"uid": uid}
	// u.ServeJSON()
}
