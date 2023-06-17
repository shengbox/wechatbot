package gpt

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"

	"github.com/869413421/wechatbot/config"
	"github.com/869413421/wechatbot/functions"
	openai "github.com/sashabaranov/go-openai"
)

const BASEURL = "https://api.openai.com/v1/"

var funcs []*openai.FunctionDefine

func init() {
	data, err := os.ReadFile("functions.json")
	fmt.Println(string(data))
	if err != nil {
		log.Fatal("Error reading file:", err)
		return
	}
	json.Unmarshal(data, &funcs)
}

// gpt-3.5-turbo
func Completions3Dot5(messages []openai.ChatCompletionMessage) (string, error) {
	log.Println("request:", messages)
	apiKey := config.LoadConfig().ApiKey
	cfg := openai.DefaultConfig(apiKey)
	cfg.BaseURL = config.LoadConfig().BaseURL
	client := openai.NewClientWithConfig(cfg)

	resp, err := client.CreateChatCompletion(
		context.Background(),
		openai.ChatCompletionRequest{
			Model:           openai.GPT3Dot5Turbo16K0613,
			Messages:        messages,
			Temperature:     1,
			MaxTokens:       2000,
			PresencePenalty: 0,
		},
	)

	if err != nil {
		fmt.Printf("ChatCompletion error: %v\n", err)
		return "", err
	}
	log.Println("resp:", resp)
	return resp.Choices[0].Message.Content, nil
}

func CreateChatCompletion(messages []openai.ChatCompletionMessage) (string, error) {
	apiKey := config.LoadConfig().ApiKey
	cfg := openai.DefaultConfig(apiKey)
	cfg.BaseURL = config.LoadConfig().BaseURL
	client := openai.NewClientWithConfig(cfg)

	resp, err := client.CreateChatCompletion(
		context.Background(),
		openai.ChatCompletionRequest{
			Model:     openai.GPT3Dot5Turbo16K0613,
			Messages:  messages,
			Functions: funcs,
		},
	)
	if err != nil {
		fmt.Printf("ChatCompletion error: %v\n", err)
		return "", err
	}
	if resp.Choices[0].FinishReason == "function_call" {
		functionCall := resp.Choices[0].Message.FunctionCall
		log.Println(functionCall.Name, functionCall.Arguments)
		var arguments map[string]string
		json.Unmarshal([]byte(functionCall.Arguments), &arguments)

		switch functionCall.Name {
		case "get_job_list":
			body := functions.GetJobList(arguments)
			messages = append(messages, openai.ChatCompletionMessage{
				Role:    openai.ChatMessageRoleFunction,
				Content: body,
				Name:    functionCall.Name,
			})
			return CreateChatCompletion(messages)
		case "get_user_info":
			body := functions.GetUserInfo(arguments)
			messages = append(messages, openai.ChatCompletionMessage{
				Role:    openai.ChatMessageRoleFunction,
				Content: body,
				Name:    functionCall.Name,
			})
			return CreateChatCompletion(messages)
		case "get_apply_list":
			body := functions.GetApplyList(arguments)
			messages = append(messages, openai.ChatCompletionMessage{
				Role:    openai.ChatMessageRoleFunction,
				Content: body,
				Name:    functionCall.Name,
			})
			return CreateChatCompletion(messages)
		default:
			return "", errors.New("unknow functionCall")
		}
	}
	return resp.Choices[0].Message.Content, nil
}
