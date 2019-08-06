package middleware

import (
	"century/oasis/api/db"
	"fmt"
	"net/http"

	ct "context"

	"github.com/astaxie/beego/context"
	"go.mongodb.org/mongo-driver/bson"
)

var FilterUser = func(ctx *context.Context) {

	fmt.Println(ctx.Request.RequestURI)
	if ctx.Request.URL.Path != `/api/v1/account` { // 路由正则过滤
		apiKey := ""
		apiKey = ctx.Input.Query("apiKey")
		if apiKey == "" {
			res := map[string]interface{}{
				"code": 40000,
				"Msg":  "not find apiKey param",
			}
			ctx.Output.SetStatus(http.StatusUnauthorized)
			ctx.Output.JSON(res, false, false)
			return
		}
		dbs := db.GetDB()
		ctx.Input.SetData("db", dbs)

		result := dbs.ConnCollection("accounts").FindOne(ct.Background(), bson.M{"apiKey": apiKey})

		if result.Err() != nil {
			res := map[string]interface{}{
				"code": 40001,
				"msg":  result.Err().Error(),
			}
			ctx.Output.SetStatus(http.StatusUnauthorized)
			ctx.Output.JSON(res, false, false)
			return
		}

		var user map[string]interface{}

		result.Decode(&user)

		if user["apiKey"].(string) == "" {
			res := map[string]interface{}{
				"code": 40001,
				"msg":  "not find apiKey in db",
			}
			ctx.Output.SetStatus(http.StatusUnauthorized)
			ctx.Output.JSON(res, false, false)
			return
		}

		ctx.Input.SetData("userId", user["_id"])

		// rawByte, _ := result.DecodeBytes()

		// raw := rawByte.Lookup("apiKey").String() // 坑 返回的带 “ " ” ，单纯字符串会有 /""/ 应去掉
		// raw = strings.TrimSuffix(raw, "\"")
		// if raw == "" {
		// 	res := map[string]interface{}{
		// 		"code": 40001,
		// 		"msg":  "not find apiKey in db",
		// 	}
		// 	ctx.Output.SetStatus(http.StatusUnauthorized)
		// 	ctx.Output.JSON(res, false, false)
		// 	return
		// }

		// rawid := rawByte.Lookup("_id")
		// if id, ok := rawid.ObjectIDOK(); ok {
		// 	ctx.Input.SetData("userId", id.Hex())
		// }
	}

}
