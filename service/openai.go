package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
	"time"

	"github.com/go-resty/resty/v2"
	_ "github.com/joho/godotenv/autoload"
	openai "github.com/sashabaranov/go-openai"
)

type FunctionCall struct {
	Name   string `json:"name"`
	API    string `json:"api"`
	Method string `json:"method"`
}

var (
	tools   []openai.Tool
	client  *openai.Client
	callMap = map[string]FunctionCall{}
)

func init() {
	cfg := openai.DefaultConfig(os.Getenv("api_key"))
	cfg.BaseURL = os.Getenv("base_URL")
	client = openai.NewClientWithConfig(cfg)

	data, err := os.ReadFile("functions.json")
	if err != nil {
		log.Println("warning reading file:", err)
		return
	}
	var funcs []openai.FunctionDefinition
	json.Unmarshal(data, &funcs)
	for _, fun := range funcs {
		tools = append(tools, openai.Tool{
			Type:     openai.ToolTypeFunction,
			Function: &fun,
		})
	}
	var calls []FunctionCall
	json.Unmarshal(data, &calls)
	for _, it := range calls {
		callMap[it.Name] = it
	}
}

func CreateChatCompletion(messages []openai.ChatCompletionMessage) (string, error) {
	resp, err := client.CreateChatCompletion(context.Background(),
		openai.ChatCompletionRequest{
			Model:    openai.GPT4oMini20240718,
			Messages: messages,
			Tools:    tools,
		},
	)
	if err != nil {
		fmt.Printf("ChatCompletion error: %v\n", err)
		return "", err
	}
	if resp.Choices[0].FinishReason == "function_call" {
		functionCall := resp.Choices[0].Message.FunctionCall
		log.Println("function_call")
		body, _ := Call(functionCall)
		messages = append(messages, openai.ChatCompletionMessage{
			Role:    openai.ChatMessageRoleFunction,
			Content: body,
			Name:    functionCall.Name,
		})
		return CreateChatCompletion(messages)
	}
	if resp.Choices[0].FinishReason == "tool_calls" {
		functionCall := resp.Choices[0].Message.ToolCalls[0].Function
		log.Println("tool_calls")
		body, _ := Call(&functionCall)
		message := openai.ChatCompletionMessage{
			Role:       openai.ChatMessageRoleTool,
			Content:    body,
			ToolCallID: resp.Choices[0].Message.ToolCalls[0].ID,
			Name:       functionCall.Name,
		}
		if strings.ContainsAny(strings.ToLower(resp.Model), "qwen") {
			message.Role = openai.ChatMessageRoleTool
		} else {
			message.Role = openai.ChatMessageRoleFunction
		}
		messages = append(messages, message)
		return CreateChatCompletion(messages)
	}
	return resp.Choices[0].Message.Content, nil
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

// AssistantCompletion
func AssistantCompletion(messages []openai.ChatCompletionMessage) (string, error) {
	threadMessages := []openai.ThreadMessage{{
		Role:    "user",
		Content: messages[len(messages)-1].Content,
	}}
	run, err := client.CreateThreadAndRun(context.Background(), openai.CreateThreadAndRunRequest{
		RunRequest: openai.RunRequest{
			AssistantID: os.Getenv("assistant_id"),
		},
		Thread: openai.ThreadRequest{Messages: threadMessages},
	})
	if err != nil {
		return "", err
	}
	for run.Status != openai.RunStatusCompleted {
		run, err = client.RetrieveRun(context.Background(), run.ThreadID, run.ID)
		if err != nil {
			return "", err
		}
		if run.Status == openai.RunStatusFailed {
			log.Println(run.LastError.Message)
			return "", errors.New(string(run.LastError.Code))
		}
		time.Sleep(time.Second)
	}

	msgs, err := client.ListMessage(context.Background(), run.ThreadID, nil, nil, nil, nil)
	if err != nil {
		return "", err
	}
	return msgs.Messages[0].Content[0].Text.Value, nil
}

// 语音转文本
func Transcription(filePath string) (string, error) {
	response, err := client.CreateTranscription(context.Background(), openai.AudioRequest{
		Model:    "whisper-1",
		FilePath: filePath,
	})
	if err != nil {
		return "", err
	}
	log.Println("voice text: ", response.Text)
	return response.Text, nil
}

// 文本转语音
func Speech(input, filename string) error {
	response, err := client.CreateSpeech(context.Background(), openai.CreateSpeechRequest{
		Model: openai.TTSModel1,
		Input: input,
		Voice: openai.VoiceAlloy,
	})
	if err != nil {
		return err
	}
	defer response.Close()
	voice, err := os.Create(filename)
	if err != nil {
		return err
	}
	_, err = io.Copy(voice, response)
	return err
}

// 函数调用
func Call(functionCall *openai.FunctionCall) (string, error) {
	call := callMap[functionCall.Name]

	var arguments map[string]string
	json.Unmarshal([]byte(functionCall.Arguments), &arguments)
	log.Println("api", call.API, arguments)

	if strings.ToUpper(call.Method) == "GET" {
		resp, err := resty.New().R().SetQueryParams(arguments).Get(call.API)
		log.Println("api result", resp.String(), err)
		return resp.String(), err
	} else {
		resp, err := resty.New().R().SetBody(arguments).Get(call.API)
		log.Println("api result", resp.String(), err)
		return resp.String(), err
	}
}
