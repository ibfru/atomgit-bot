package routers

import (
	"sig-info-cache/controllers"

	"github.com/beego/beego/v2/server/web"
)

func init() {
	web.Router("/", &controllers.MainController{})
	web.Router("/sig/:org/:repo", &controllers.SigInfoController{}, "get:GetSigInfo")
	//web.Include(&controllers.SigInfoController{})
}
