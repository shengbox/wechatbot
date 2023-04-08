package test

import (
	"context"
	"errors"
	"fmt"
	"io"
	"testing"

	"github.com/869413421/wechatbot/config"
	"github.com/sashabaranov/go-openai"
)

func TestGptStream(t *testing.T) {
	apiKey := config.LoadConfig().ApiKey
	cfg := openai.DefaultConfig(apiKey)
	cfg.BaseURL = config.LoadConfig().BaseURL
	client := openai.NewClientWithConfig(cfg)

	messages := []openai.ChatCompletionMessage{
		{Role: openai.ChatMessageRoleUser, Content: "你好呀"},
	}

	req := openai.ChatCompletionRequest{
		Model:           openai.GPT3Dot5Turbo,
		Messages:        messages,
		Temperature:     1,
		MaxTokens:       2000,
		PresencePenalty: 0,
		Stream:          true,
	}
	stream, err := client.CreateChatCompletionStream(context.Background(), req)
	if err != nil {
		fmt.Printf("ChatCompletionStream error: %v\n", err)
		return
	}
	defer stream.Close()

	fmt.Printf("Stream response: ")
	for {
		response, err := stream.Recv()
		if errors.Is(err, io.EOF) {
			fmt.Println("\nStream finished")
			return
		}
		if err != nil {
			fmt.Printf("\nStream error: %v\n", err)
			return
		}
		fmt.Printf(response.Choices[0].Delta.Content)
	}
}

func TestImage(t *testing.T) {
	c := openai.NewClient("")

	// Sample image by link
	reqUrl := openai.ImageRequest{
		Prompt:         "A pig riding a skateboard, in a cartoon style.",
		Size:           openai.CreateImageSize256x256,
		ResponseFormat: openai.CreateImageResponseFormatURL,
		N:              1,
	}

	respUrl, err := c.CreateImage(context.Background(), reqUrl)
	if err != nil {
		fmt.Printf("Image creation error: %v\n", err)
		return
	}
	fmt.Println(respUrl.Data[0].URL)

	// // Example image as base64
	// reqBase64 := openai.ImageRequest{
	// 	Prompt:         "Portrait of a humanoid parrot in a classic costume, high detail, realistic light, unreal engine",
	// 	Size:           openai.CreateImageSize256x256,
	// 	ResponseFormat: openai.CreateImageResponseFormatB64JSON,
	// 	N:              1,
	// }

	// respBase64, err := c.CreateImage(ctx, reqBase64)
	// if err != nil {
	// 	fmt.Printf("Image creation error: %v\n", err)
	// 	return
	// }

	// imgBytes, err := base64.StdEncoding.DecodeString(respBase64.Data[0].B64JSON)
	// if err != nil {
	// 	fmt.Printf("Base64 decode error: %v\n", err)
	// 	return
	// }

	// r := bytes.NewReader(imgBytes)
	// imgData, err := png.Decode(r)
	// if err != nil {
	// 	fmt.Printf("PNG decode error: %v\n", err)
	// 	return
	// }

	// file, err := os.Create("example.png")
	// if err != nil {
	// 	fmt.Printf("File creation error: %v\n", err)
	// 	return
	// }
	// defer file.Close()

	// if err := png.Encode(file, imgData); err != nil {
	// 	fmt.Printf("PNG encode error: %v\n", err)
	// 	return
	// }

	// fmt.Println("The image was saved as example.png")
}
