package controllers

import (
	"fmt"

	"github.com/beego/beego/v2/server/web"
)

type MainController struct {
	web.Controller
}

func (c *MainController) Get() {
	c.Data["Website"] = "beego.vip"
	c.Data["Email"] = "astaxie@gmail.com"
	c.TplName = "index.tpl"
}

type SigInfoController struct {
	web.Controller
}

func (s *SigInfoController) URLMapping() {
	s.Mapping("GetSigInfo", s.GetSigInfo)
}

func (s *SigInfoController) GetSigInfo() {
	fmt.Println(s.GetString(":org"), s.GetString(":repo"))
	s.Ctx.WriteString("111111111111111")
}
