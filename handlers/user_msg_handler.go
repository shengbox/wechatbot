package handlers

import (
	"fmt"
	"log"
	"strings"

	"github.com/869413421/wechatbot/gtp"
	"github.com/eatmoreapple/openwechat"
	"github.com/sashabaranov/go-openai"
)

var _ MessageHandlerInterface = (*UserMessageHandler)(nil)

// UserMessageHandler 私聊消息处理
type UserMessageHandler struct {
}

// handle 处理消息
func (g *UserMessageHandler) handle(msg *openwechat.Message) error {
	if msg.IsText() {
		return g.ReplyText(msg)
	}
	return nil
}

// NewUserMessageHandler 创建私聊处理器
func NewUserMessageHandler() MessageHandlerInterface {
	return &UserMessageHandler{}
}

// ReplyText 发送文本消息到群
func (g *UserMessageHandler) ReplyText(msg *openwechat.Message) error {
	// 接收私聊消息
	sender, err := msg.Sender()
	if err != nil {
		fmt.Println(err)
	}
	log.Printf("Received User %v Text Msg : %v", sender.NickName, msg.Content)
	if UserService.ClearUserSessionContext(sender.ID(), msg.Content) {
		_, err = msg.ReplyText("上下文已经清空了，你可以问下一个问题啦。")
		if err != nil {
			log.Printf("response user error: %v \n", err)
		}
		return nil
	}

	// 获取上下文，向GPT发起请求
	requestText := strings.TrimSpace(msg.Content)
	requestText = strings.Trim(requestText, "\n")

	messages := UserService.GetUserSessionContext(sender.ID())
	messages = append(messages, openai.ChatCompletionMessage{
		Role:    openai.ChatMessageRoleUser,
		Content: requestText,
	})
	// 保留5个上下文
	if len(messages) > 10 {
		messages = messages[len(messages)-10:]
		messages = append(messages, openai.ChatCompletionMessage{
			Role:    openai.ChatMessageRoleSystem,
			Content: "You are a helpful assistant.",
		})
	}
	reply, err := gtp.Completions3Dot5(messages)
	if err != nil {
		log.Printf("gtp request error: %v \n", err)
		msg.ReplyText("机器人神了，我一会发现了就去修。")
		return err
	}
	if reply == "" {
		return nil
	}

	// 设置上下文，回复用户
	reply = strings.TrimSpace(reply)
	reply = strings.Trim(reply, "\n")
	messages = append(messages, openai.ChatCompletionMessage{
		Role:    openai.ChatMessageRoleAssistant,
		Content: reply,
	})
	UserService.SetUserSessionContext(sender.ID(), messages)
	reply = "本消息由 chatGPT 回复：\n" + reply
	_, err = msg.ReplyText(reply)
	if err != nil {
		log.Printf("response user error: %v \n", err)
	}
	return err
}
