package routers

import (
	"github.com/kapacitor_ui/controllers"
	"github.com/astaxie/beego"
)

func init() {
    beego.Router("/", &controllers.MainController{})
}
