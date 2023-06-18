package gpt

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/869413421/wechatbot/functions"
	_ "github.com/joho/godotenv/autoload"
	openai "github.com/sashabaranov/go-openai"
)

var (
	funcs  []*openai.FunctionDefine
	client *openai.Client
)

func init() {
	data, err := os.ReadFile("functions.json")
	if err != nil {
		log.Fatal("Error reading file:", err)
	}
	json.Unmarshal(data, &funcs)

	cfg := openai.DefaultConfig(os.Getenv("api_key"))
	cfg.BaseURL = os.Getenv("base_URL")
	client = openai.NewClientWithConfig(cfg)
}

// gpt-3.5-turbo
func Completions3Dot5(messages []openai.ChatCompletionMessage) (string, error) {
	log.Println("request:", messages)
	resp, err := client.CreateChatCompletion(
		context.Background(),
		openai.ChatCompletionRequest{
			Model:           openai.GPT3Dot5Turbo16K0613,
			Messages:        messages,
			Temperature:     1,
			MaxTokens:       4000,
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

		body, _ := functions.Call(functionCall.Name, arguments)
		messages = append(messages, openai.ChatCompletionMessage{
			Role:    openai.ChatMessageRoleFunction,
			Content: body,
			Name:    functionCall.Name,
		})
		return CreateChatCompletion(messages)
	}
	return resp.Choices[0].Message.Content, nil
}
