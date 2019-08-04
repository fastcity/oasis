package middleware

import (
	"century/oasis/api/db"
	"fmt"

	ct "context"

	"github.com/astaxie/beego/context"
	"go.mongodb.org/mongo-driver/bson"
)

var FilterUser = func(ctx *context.Context) {
	// _, ok := ctx.Input.Session("uid").(int)
	// if !ok && ctx.Request.RequestURI != "/login" {
	// 	ctx.Redirect(302, "/login")
	// }
	dbs := db.Init()
	ctx.Input.SetData("db", dbs)
	apiKey := ctx.Input.Query("apiKey")
	if apiKey == "" {
		res := map[string]interface{}{
			"code": 40000,
			"Msg":  "not find apiKey",
		}

		ctx.Output.JSON(res, false, false)
		return
	}

	result := dbs.ConnCollection("accounts").FindOne(ct.Background(), bson.M{"apiKey": apiKey})
	fmt.Println(result)
	if result.Err() != nil {
		res := map[string]interface{}{
			"code": 40001,
			"msg":  result.Err().Error(),
		}
		ctx.Output.JSON(res, false, false)
		return
	}

}
