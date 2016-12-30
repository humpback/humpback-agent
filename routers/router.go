package routers

import (
	"humpback-agent/controllers"
	"humpback-agent/validators"

	"github.com/astaxie/beego"
)

// Init - Init routers
func Init() {
	faqRouter := beego.NSRouter("/humpback-agent-faq", &controllers.FaqController{})

	imageRouters := beego.NSNamespace("/images",
		beego.NSRouter("/", &controllers.ImageController{}, "get:GetImages;post:PullImage"),
		beego.NSRouter("/*", &controllers.ImageController{}, "get:GetImage;delete:DeleteImage"),
	)

	containerRouters := beego.NSNamespace("/containers",
		beego.NSRouter("/", &controllers.ContainerController{}, "get:GetContainers;post:CreateContainer;put:OperateContainer"),
		beego.NSRouter("/:containerid", &controllers.ContainerController{}, "get:GetContainer;delete:DeleteContainer"),
		beego.NSRouter("/:containerid/logs", &controllers.ContainerController{}, "get:GetContainerLogs"),
		beego.NSRouter("/:containerid/stats", &controllers.ContainerController{}, "get:GetContainerStats"),
	)

	agentSpace := beego.NewNamespace("/v1",
		faqRouter,
		imageRouters,
		containerRouters,
	)
	beego.AddNamespace(agentSpace)

	beego.InsertFilter("/v1/containers", beego.BeforeExec, validators.CreateContainerValidator)
}
