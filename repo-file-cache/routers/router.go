package routers

import (
	beego "github.com/beego/beego/v2/server/web"

	"github.com/opensourceways/repo-file-cache/controllers"
)

func init() {
	ns := beego.NewNamespace("/v1",
		beego.NSNamespace("/file",
			beego.NSInclude(
				&controllers.FileController{},
			),
		),
	)
	beego.AddNamespace(ns)
}
