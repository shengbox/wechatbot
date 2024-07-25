package gpt

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"time"

	"github.com/869413421/wechatbot/functions"
	_ "github.com/joho/godotenv/autoload"
	openai "github.com/sashabaranov/go-openai"
)

var (
	funcs  []openai.FunctionDefinition
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
			Model:     openai.GPT4oMini20240718,
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
