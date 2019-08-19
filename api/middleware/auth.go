package middleware

import (
	"century/oasis/api/db"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"net/http"
	"sort"
	"strings"

	"github.com/astaxie/beego"

	ct "context"

	"github.com/astaxie/beego/context"
	"go.mongodb.org/mongo-driver/bson"
)

const (
	defaultMaxMemory = 32 << 20 // 32 MB
)

func sortKeys(data map[string][]string) []string {
	keys := getKeys(data)
	sort.Strings(keys)
	return keys

}

func getKeys(data map[string][]string) []string {
	keys := []string{}
	for k, _ := range data {
		keys = append(keys, k)
	}
	return keys
}

func sign(ctx *context.Context) string {
	// form := map[string]string{}
	ctx.Request.ParseMultipartForm(defaultMaxMemory)
	body := ctx.Request.Form

	sortKey := sortKeys(body)

	data := ""
	for _, v := range sortKey {
		// 数组参数
		if v != "signature" {
			sort.Strings(body[v])
			for _, fv := range body[v] {
				if fv != "" {
					data += v + "=" + fv + "&"
				}
			}
		}
	}
	data = strings.TrimRight(data, "&")
	fmt.Println("sign data", data)
	beego.Debug("sign data", data)

	h := md5.New()
	h.Write([]byte(data))
	cipherStr := h.Sum(nil)

	digest := hex.EncodeToString(cipherStr)

	// ctx.Set("signature", digest)
	// fmt.Printf("%s\n", digest) // 输出加密结果
	beego.Debug(digest)

	return strings.ToLower(digest)
}

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

		///api/v1/balance
		// /api/v1/balance?address=ARJtq6Q46oTnxDwvVqMgDtZeNxs7Ybt81A&apiKey=12345678&signature=c191872682d3ad1ac0ad6ff71c3fc62b
		// 5221ef283de4018031bbde93b1a0aa37
		si := sign(ctx)
		signature := ctx.Input.Query("signature")
		if strings.Compare(si, strings.ToLower(signature)) != 0 {
			res := map[string]interface{}{
				"code": 40001,
				"msg":  "signature not match,show be" + si + "get" + signature,
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

		// switch str := user["apiKey"].(type) {
		// case string:
		// 	if str == "" {
		// 		res := map[string]interface{}{
		// 			"code": 40001,
		// 			"msg":  "not find apiKey in db",
		// 		}
		// 		ctx.Output.SetStatus(http.StatusUnauthorized)
		// 		ctx.Output.JSON(res, false, false)
		// 		return
		// 	}

		// }
		if _, ok := user["apiKey"]; !ok || user["apiKey"].(string) == "" {
			res := map[string]interface{}{
				"code": 40001,
				"msg":  "not find apiKey in db",
			}
			ctx.Output.SetStatus(http.StatusUnauthorized)
			ctx.Output.JSON(res, false, false)
			return
		}

		// if user["apiKey"].(string) == "" {
		// 	res := map[string]interface{}{
		// 		"code": 40001,
		// 		"msg":  "not find apiKey in db",
		// 	}
		// 	ctx.Output.SetStatus(http.StatusUnauthorized)
		// 	ctx.Output.JSON(res, false, false)
		// 	return
		// }

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
