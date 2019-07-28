package controllers

import (
	"century/api/models"
	"encoding/json"
	"fmt"

	"github.com/astaxie/beego"
	"github.com/astaxie/beego/httplib"
)

// Operations about Users
type BalanceController struct {
	beego.Controller
}

func (u *BalanceController) Get() {
	var balance models.Balance
	json.Unmarshal(u.Ctx.Input.RequestBody, &balance)
	host := beego.AppConfig.String("builderHost")
	port := beego.AppConfig.String("builderPort")
	url := fmt.Sprintf("http://%s:%s/api/getBalance", host, port)
	req := httplib.Get(url)
	var resp models.CommResp
	fmt.Println(req.Bytes())
	err := req.ToJSON(&resp)
	if err != nil {
		u.Data["json"] = map[string]interface{}{
			"code": 40000,
			"msg":  err.Error(),
		}

		u.ServeJSON()
		return
	}
	fmt.Println(resp)
	u.Data["json"] = map[string]interface{}{
		"code": 0,
		"data": resp.Data,
	}
	u.ServeJSON()
	// resp, err := http.Get(url)
	// if err != nil {

	// }
	// resp.Body
	// uid := models.AddUser(user)
	// u.Data["json"] = map[string]string{"uid": uid}
	// u.ServeJSON()
}
