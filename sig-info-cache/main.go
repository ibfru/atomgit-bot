package main

import (
	"fmt"
	_ "sig-info-cache/routers"

	"github.com/beego/beego/v2/server/web"
)

func main() {
	//err := web.LoadAppConfig("ini", "D:\\Project\\github\\ibfru\\atomgit-bot\\sig-info-cache\\conf\\app.conf")
	//if err != nil {
	//	return
	//}
	configFile, err := web.AppConfig.String("appconf")
	if err != nil {
		_ = fmt.Errorf("error %+v", err)
	}
	fmt.Println(configFile)

	web.Run()
}
