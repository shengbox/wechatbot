package handlers

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/869413421/wechatbot/service"
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
	if msg.IsVoice() {
		mp3file := msg.MsgId + ".mp3"
		msg.SaveFileToLocal(mp3file)
		txt, err := service.Transcription(mp3file)
		// txt, err := funasr.SpeechToText(mp3file, nil)
		os.Remove(mp3file)
		if err != nil {
			log.Printf("gtp request error: %v \n", err)
			msg.ReplyText("不知道发生了什么，我一会发现了就去修。")
			return err
		}
		msg.Content = txt
		return g.ReplyText(msg)
	}
	if msg.IsPicture() {
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

	messages := UserService.GetUserSessionContext(sender.ID(), sender.NickName)
	switch msg.MsgType {
	case openwechat.MsgTypeImage:
		log.Println("收到一张图片")
		buffer := &bytes.Buffer{}
		_ = msg.SaveFile(buffer)
		encoded := base64.StdEncoding.EncodeToString(buffer.Bytes())
		messages = append(messages, openai.ChatCompletionMessage{
			Role: openai.ChatMessageRoleUser,
			MultiContent: []openai.ChatMessagePart{{
				Type: openai.ChatMessagePartTypeImageURL,
				ImageURL: &openai.ChatMessageImageURL{
					URL: fmt.Sprintf("data:image/jpeg;base64,%s", encoded),
				},
			}},
		})
	default:
		// 获取上下文，向GPT发起请求
		requestText := strings.TrimSpace(msg.Content)
		requestText = strings.Trim(requestText, "\n")
		messages = append(messages, openai.ChatCompletionMessage{
			Role:    openai.ChatMessageRoleUser,
			Content: requestText,
		})
	}

	// 保留5个上下文
	// if len(messages) > 10 {
	// 	messages = messages[len(messages)-10:]
	// 	if os.Getenv("prompt.system") != "" {
	// 		messages = append(messages, openai.ChatCompletionMessage{
	// 			Role:    openai.ChatMessageRoleSystem,
	// 			Content: os.Getenv("prompt.system"),
	// 		})
	// 	}
	// }
	var reply string
	if os.Getenv("assistant_id") != "" {
		reply, err = service.AssistantCompletion(messages)
	} else {
		reply, err = service.CreateChatCompletion(messages)
	}
	if err != nil {
		log.Printf("gtp request error: %v \n", err)
		msg.ReplyText("不知道发生了什么，我一会发现了就去修。")
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
	// reply = "数字助理：\n" + reply

	switch msg.MsgType {
	case openwechat.MsgTypeVideo: // 暂时移除tts服务
		err = service.Speech(reply, msg.MsgId+".mp3")
		if err != nil {
			return err
		}
		voice, _ := os.Open(msg.MsgId + ".mp3")
		_, err = msg.ReplyFile(voice)
		os.Remove(msg.MsgId + ".mp3")
	default:
		_, err = msg.ReplyText(reply)
	}
	// _, err = msg.ReplyText(reply)
	if err != nil {
		log.Printf("response user error: %v \n", err)
	}
	return err
}
