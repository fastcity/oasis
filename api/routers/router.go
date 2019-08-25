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

	dbs := db.GetDB()

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

	ft := &controllers.TransferController{DB: dbs}
	middle := middleware.NewMiddle(dbs)

	nsAcc := beego.NewNamespace("/api/v1",
		// beego.NSBefore(middleware.FilterUser),
		beego.NSRouter("/balance", &controllers.BalanceController{}),
		beego.NSRouter("/createTransferTxData", ft, "post:CreateTransferTxData"),
		beego.NSRouter("/submitTx", ft, "post:SubmitTx"),

		beego.NSRouter("/getTxStatus", ft, "get:GetTxStatus"),

		beego.NSRouter("/subscribe", &controllers.AccountController{DB: dbs}, "post:Subscribe"),

		beego.NSNamespace("/account",
			beego.NSInclude(
				&controllers.AccountController{DB: dbs},
			),
		),
	)
	beego.AddNamespace(nsAcc)

	//TODO: spacve 需要有正则路由，示例
	// beego.Any("/*", func(ctx *bctx.Context) {
	// 	ctx.Output.SetStatus(http.StatusNotFound)
	// 	ctx.Output.JSON(map[string]interface{}{
	// 		"code": 404,
	// 		"msg":  "unsupported rotuer",
	// 	}, false, false)
	// })
	beego.InsertFilter("/*", beego.BeforeRouter, middle.Auth()) // TODO:  /api/v1/account 不用过滤 正则
}
