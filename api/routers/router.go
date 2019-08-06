// @APIVersion 1.0.0
// @Title beego Test API
// @Description beego has a very cool tools to autogenerate documents for your API
// @Contact astaxie@gmail.com
// @TermsOfServiceUrl http://beego.me/
// @License Apache 2.0
// @LicenseUrl http://www.apache.org/licenses/LICENSE-2.0.html
package routers

import (
	"century/oasis/api/controllers"
	"century/oasis/api/db"
	"century/oasis/api/middleware"

	"github.com/astaxie/beego"
)

func init() {

	db := db.GetDB()

	ns := beego.NewNamespace("/v1",
		beego.NSNamespace("/object",
			beego.NSInclude(
				&controllers.ObjectController{},
			),
		),
		beego.NSNamespace("/user",
			beego.NSInclude(
				&controllers.UserController{},
			),
		),
	)
	beego.AddNamespace(ns)

	ft := &controllers.TransferController{DB: db}
	// beego.Router("/", &controllers.MainController{})
	// beego.Router("api/v1/", &controllers.TransferController{}, "post:CreateTransferTxData")
	// beego.Router("api/v1/balance", &controllers.BalanceController{})
	// beego.Router("api/v1/createTransferTxData", ft, "post:CreateTransferTxData")
	// beego.Router("api/v1/submitTx", ft, "post:SubmitTx")

	// beego.Router("api/v1/account", &controllers.AccountController{DB: db})

	nsAcc := beego.NewNamespace("/api/v1",
		// beego.NSBefore(middleware.FilterUser),
		beego.NSRouter("/balance", &controllers.BalanceController{}),
		beego.NSRouter("/createTransferTxData", ft, "post:CreateTransferTxData"),
		beego.NSRouter("/submitTx", ft, "post:SubmitTx"),

		beego.NSRouter("/getTxStatus", ft, "get:GetTxStatus"),

		beego.NSRouter("/subscribe", &controllers.AccountController{DB: db}, "post:Subscribe"),

		beego.NSNamespace("/account",
			beego.NSInclude(
				&controllers.AccountController{DB: db},
			),
		),
	)
	beego.AddNamespace(nsAcc)
	beego.InsertFilter("/*", beego.BeforeRouter, middleware.FilterUser) // TODO:  /api/v1/account 不用过滤 正则
}
