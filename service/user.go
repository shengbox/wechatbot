package service

import (
	"strings"
	"time"
	"unicode/utf8"

	"github.com/869413421/wechatbot/config"
	"github.com/patrickmn/go-cache"
	"github.com/sashabaranov/go-openai"
)

// UserServiceInterface 用户业务接口
type UserServiceInterface interface {
	GetUserSessionContext(userId string) []openai.ChatCompletionMessage
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
	return &UserService{cache: cache.New(time.Second*config.LoadConfig().SessionTimeout, time.Minute*10)}
}

// GetUserSessionContext 获取用户会话上下文文本
func (s *UserService) GetUserSessionContext(userId string) []openai.ChatCompletionMessage {
	sessionContext, ok := s.cache.Get(userId)
	if !ok {
		return []openai.ChatCompletionMessage{}
	}
	return sessionContext.([]openai.ChatCompletionMessage)
}

// SetUserSessionContext 设置用户会话上下文文本，question用户提问内容，GTP回复内容
func (s *UserService) SetUserSessionContext(userId string, value []openai.ChatCompletionMessage) {
	// value := question + "\n" + reply
	s.cache.Set(userId, value, time.Second*config.LoadConfig().SessionTimeout)
}
