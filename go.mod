module github.com/869413421/wechatbot

go 1.20

require (
	github.com/eatmoreapple/openwechat v1.4.3
	github.com/joho/godotenv v1.5.1
	github.com/patrickmn/go-cache v2.1.0+incompatible
	github.com/sashabaranov/go-openai v1.24.0
	github.com/shengbox/go-util v0.0.0-20240522070802-8360992a63ba
	github.com/skip2/go-qrcode v0.0.0-20200617195104-da1b6568686e
)

require (
	github.com/gorilla/websocket v1.5.1 // indirect
	golang.org/x/net v0.17.0 // indirect
)

// replace github.com/sashabaranov/go-openai => ../go-openai
