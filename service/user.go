package service

import (
	"fmt"
	"os"
	"strings"
	"time"
	"unicode/utf8"

	_ "github.com/joho/godotenv/autoload"
	"github.com/patrickmn/go-cache"
	"github.com/sashabaranov/go-openai"
)

// UserServiceInterface 用户业务接口
type UserServiceInterface interface {
	GetUserSessionContext(userId, nickname string) []openai.ChatCompletionMessage
	SetUserSessionContext(userId string, messages []openai.ChatCompletionMessage)
	ClearUserSessionContext(userId string, msg string) bool
}

var _ UserServiceInterface = (*UserService)(nil)

// UserService 用戶业务
type UserService struct {
	// 缓存
	cache *cache.Cache
}

// ClearUserSessionContext 清空GTP上下文，接收文本中包含`我要问下一个问题`，并且Unicode 字符数量不超过20就清空
func (s *UserService) ClearUserSessionContext(userId string, msg string) bool {
	if strings.Contains(msg, "我要问下一个问题") && utf8.RuneCountInString(msg) < 20 {
		s.cache.Delete(userId)
		return true
	}
	return false
}

// NewUserService 创建新的业务层
func NewUserService() UserServiceInterface {
	sessionTimeout, _ := time.ParseDuration(os.Getenv("session_timeout"))
	return &UserService{cache: cache.New(sessionTimeout, time.Minute*10)}
}

// GetUserSessionContext 获取用户会话上下文文本
func (s *UserService) GetUserSessionContext(userId, nickname string) []openai.ChatCompletionMessage {
	sessionContext, ok := s.cache.Get(userId)
	if !ok {
		if os.Getenv("prompt.system") == "" {
			return []openai.ChatCompletionMessage{}
		}
		return []openai.ChatCompletionMessage{
			{
				Role:    openai.ChatMessageRoleSystem,
				Content: fmt.Sprintf("%s,你在与你的用户对话，用户的昵称是:%s,今天是%s", os.Getenv("prompt.system"), nickname, time.Now().Format("2006年01月02日")),
			},
		}
	}
	return sessionContext.([]openai.ChatCompletionMessage)
}

// SetUserSessionContext 设置用户会话上下文文本，question用户提问内容，GTP回复内容
func (s *UserService) SetUserSessionContext(userId string, value []openai.ChatCompletionMessage) {
	// value := question + "\n" + reply
	sessionTimeout, _ := time.ParseDuration(os.Getenv("session_timeout"))
	s.cache.Set(userId, value, sessionTimeout)
}
