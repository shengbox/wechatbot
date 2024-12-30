package handlers

import (
	"log"
	"os"
	"strings"

	"github.com/869413421/wechatbot/service"
	"github.com/eatmoreapple/openwechat"
	"github.com/sashabaranov/go-openai"
)

var _ MessageHandlerInterface = (*GroupMessageHandler)(nil)

// GroupMessageHandler 群消息处理
type GroupMessageHandler struct {
}

// handle 处理消息
func (g *GroupMessageHandler) handle(msg *openwechat.Message) error {
	if msg.IsText() {
		return g.ReplyText(msg)
	}
	return nil
}

// NewGroupMessageHandler 创建群消息处理器
func NewGroupMessageHandler() MessageHandlerInterface {
	return &GroupMessageHandler{}
}

// ReplyText 发送文本消息到群
func (g *GroupMessageHandler) ReplyText(msg *openwechat.Message) error {
	// 接收群消息
	sender, err := msg.Sender()
	group := openwechat.Group{User: sender}
	log.Printf("Received Group %v Text Msg : %v", group.NickName, msg.Content)

	// 不是@的不处理
	if !msg.IsAt() {
		return nil
	}

	// 获取@我的用户
	groupSender, err := msg.SenderInGroup()
	if err != nil {
		log.Printf("get sender in group error :%v \n", err)
		return err
	}
	atText := "@" + groupSender.NickName + " "

	if UserService.ClearUserSessionContext(sender.ID(), msg.Content) {
		_, err = msg.ReplyText(atText + "上下文已经清空了，你可以问下一个问题啦。")
		if err != nil {
			log.Printf("response user error: %v \n", err)
		}
		return nil
	}

	// 替换掉@文本，设置会话上下文，然后向GPT发起请求。
	messages := buildRequestText(sender, msg)
	if messages == nil {
		return nil
	}
	var reply string
	if os.Getenv("assistant_id") != "" {
		reply, err = service.AssistantCompletion(*messages)
	} else {
		reply, err = service.CreateChatCompletion(*messages)
	}
	if err != nil {
		log.Printf("gtp request error: %v \n", err)
		_, err = msg.ReplyText("不知道发生了什么，我一会发现了就去修。")
		if err != nil {
			log.Printf("response group error: %v \n", err)
		}
		return err
	}
	if reply == "" {
		return nil
	}

	// 回复@我的用户
	reply = strings.TrimSpace(reply)
	reply = strings.Trim(reply, "\n")
	// 设置上下文
	*messages = append(*messages, openai.ChatCompletionMessage{
		Role:    openai.ChatMessageRoleAssistant,
		Content: reply,
	})
	UserService.SetUserSessionContext(sender.ID(), *messages)
	replyText := atText + reply
	_, err = msg.ReplyText(replyText)
	if err != nil {
		log.Printf("response group error: %v \n", err)
	}
	return err
}

// buildRequestText 构建请求GPT的文本，替换掉机器人名称，然后检查是否有上下文，如果有拼接上
func buildRequestText(sender *openwechat.User, msg *openwechat.Message) *[]openai.ChatCompletionMessage {
	groupSender, _ := msg.SenderInGroup()

	// replaceText := "@" + sender.NickName
	replaceText := "@" + groupSender.NickName
	requestText := strings.TrimSpace(strings.ReplaceAll(msg.Content, replaceText, ""))
	if requestText == "" {
		return nil
	}
	messages := UserService.GetUserSessionContext(sender.ID()+groupSender.ID(), groupSender.NickName)
	messages = append(messages, openai.ChatCompletionMessage{
		Role:    openai.ChatMessageRoleUser,
		Content: requestText,
	})
	return &messages
}
