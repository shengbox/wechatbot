package bootstrap

import (
	"log"
	"os"

	"github.com/eatmoreapple/openwechat"
	"github.com/shengbox/wechatbot/handlers"
)

func Run() {
	//bot := openwechat.DefaultBot()
	bot := openwechat.DefaultBot(openwechat.Desktop) // 桌面模式，上面登录不上的可以尝试切换这种模式

	// 注册消息处理函数
	bot.MessageHandler = handlers.Handler

	// 注册登陆二维码回调
	bot.UUIDCallback = handlers.QrCodeCallBack

	// 创建热存储容器对象
	os.Mkdir("cache", os.ModeDir)
	reloadStorage := openwechat.NewFileHotReloadStorage("cache/storage.json")

	bot.LoginCallBack = func(body openwechat.CheckLoginResponse) {
		log.Println(string(body))
		// to do your business
		if self, err := bot.GetCurrentUser(); err == nil {
			log.Println("LoginCallBack", self.String())
		}
	}

	// 执行热登录
	err := bot.HotLogin(reloadStorage)
	if err != nil {
		if err = bot.Login(); err != nil {
			log.Printf("login error: %v \n", err)
			return
		}
	}

	// 阻塞主goroutine, 直到发生异常或者用户主动退出
	bot.Block()
}
