package routers

import (
	"github.com/astaxie/beego"
	"github.com/kapacitor_ui/controllers"
)

func init() {
	beego.Router("/", &controllers.MainController{})

	// 生成kapacitor tick
	beego.Router("/generate-tick", &controllers.MainController{}, "post:GenerateTick")
}
