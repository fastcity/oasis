package controllers

import (
	"century/api/models"
	"encoding/json"

	"github.com/astaxie/beego"
)

// Operations about Users
type TransferController struct {
	beego.Controller
}

// @Title CreateUser
// @Description create users
// @Param	body		body 	models.User	true		"body for user content"
// @Success 200 {int} models.User.Id
// @Failure 403 body is empty
// @router / [post]
func (u *TransferController) CreateTransferTxData() {
	var transfer models.Transfer
	json.Unmarshal(u.Ctx.Input.RequestBody, &transfer)
	// uid := models.AddUser(user)
	// u.Data["json"] = map[string]string{"uid": uid}
	// u.ServeJSON()
}

func (u *TransferController) getBalance() {
	var balance models.Balance
	json.Unmarshal(u.Ctx.Input.RequestBody, &balance)
	// uid := models.AddUser(user)
	// u.Data["json"] = map[string]string{"uid": uid}
	// u.ServeJSON()
}
