module github.com/869413421/wechatbot

go 1.20

require (
	github.com/eatmoreapple/openwechat v1.4.3
	github.com/go-resty/resty/v2 v2.7.0
	github.com/patrickmn/go-cache v2.1.0+incompatible
	github.com/sashabaranov/go-openai v1.11.1
	github.com/skip2/go-qrcode v0.0.0-20200617195104-da1b6568686e
)

require golang.org/x/net v0.0.0-20211029224645-99673261e6eb // indirect

// replace github.com/sashabaranov/go-openai => ../go-openai
