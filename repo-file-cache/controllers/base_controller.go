package controllers

import (
	beego "github.com/beego/beego/v2/server/web"
)

type baseController struct {
	beego.Controller
}

func (bc *baseController) apiPrepare() {
	if fr := bc.checkPathParameter(); fr != nil {
		bc.sendFailedResultAsResp(fr, "")
		bc.StopRun()
	}
}
