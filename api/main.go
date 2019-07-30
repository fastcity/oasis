package main

import (
	_ "century/oasis/api/routers"

	"github.com/astaxie/beego"
)

func main() {
	if beego.BConfig.RunMode == "dev" {
		beego.BConfig.WebConfig.DirectoryIndex = true
		beego.BConfig.WebConfig.StaticDir["/swagger"] = "swagger"
	}
	// beego.LoadAppConfig("ini", "conf/app2.conf")
	beego.Run()
}

//
