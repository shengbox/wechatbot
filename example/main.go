package main

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/go-resty/resty/v2"
	openai "github.com/sashabaranov/go-openai"
)

func main() {
	client := openai.NewClient("")
	messages := []openai.ChatCompletionMessage{
		{
			Role:    openai.ChatMessageRoleUser,
			Content: "肯德基有在招人吗？",
		},
	}
	functions := []*openai.FunctionDefine{
		{
			Name:        "get_job_list",
			Description: "获取目前正在招聘中的岗位数据",
			Parameters: &openai.FunctionParams{
				Type: openai.JSONSchemaTypeObject,
				Properties: map[string]*openai.JSONSchemaDefine{
					"brandName": {
						Type:        openai.JSONSchemaTypeString,
						Description: "在招聘岗位的商家品牌名称",
					},
					"city": {
						Type:        openai.JSONSchemaTypeString,
						Description: "岗位所在城市",
					},
				},
				Required: []string{"brandName"},
			},
		},
	}
	resp, err := client.CreateChatCompletion(
		context.Background(),
		openai.ChatCompletionRequest{
			Model:     openai.GPT3Dot5Turbo16K0613,
			Messages:  messages,
			Functions: functions,
		},
	)
	if err != nil {
		fmt.Printf("ChatCompletion error: %v\n", err)
		return
	}
	if resp.Choices[0].FinishReason == "function_call" {
		functionCallName := resp.Choices[0].Message.FunctionCall.Name
		fmt.Println(functionCallName)
		switch functionCallName {
		case "get_job_list":
			fmt.Println(resp.Choices[0].Message.FunctionCall.Arguments)
			var arguments map[string]string
			json.Unmarshal([]byte(resp.Choices[0].Message.FunctionCall.Arguments), &arguments)
			body := getJobList(arguments)

			messages = append(messages, openai.ChatCompletionMessage{
				Role:    openai.ChatMessageRoleFunction,
				Content: body,
				Name:    functionCallName,
			})
			resp, err = client.CreateChatCompletion(
				context.Background(),
				openai.ChatCompletionRequest{
					Model:     openai.GPT3Dot5Turbo16K0613,
					Messages:  messages,
					Functions: functions,
				},
			)
			if err != nil {
				fmt.Printf("ChatCompletion error: %v\n", err)
				return
			}
		}
	}
	fmt.Println(resp.Choices[0].Message.Content)
}

func getJobList(arguments map[string]string) string {
	var result Result
	resp, err := resty.New().R().
		SetResult(&result).
		SetQueryParams(arguments).
		Get("https://api.aifusheng.com/api/zhaopin/index?sort=1&current=1&pageSize=3")
	fmt.Println(resp.String(), err)
	body, _ := json.Marshal(result.Data)
	return string(body)
}

type Result struct {
	Code    int64  `json:"code"`
	Success bool   `json:"success"`
	Message string `json:"message"`
	Data    []Job  `json:"data"`
}

type Job struct {
	Name              string   `json:"name"`
	Tags              []string `json:"tags"`
	BrandName         string   `json:"brandName"`
	MoneyMin          int64    `json:"moneyMin"`
	MoneyMax          int64    `json:"moneyMax"`
	SalaryUnit        string   `json:"salaryUnit"`
	District          string   `json:"district"`
	NearbyPublicTrans string   `json:"nearbyPublicTrans"`
	JobType           string   `json:"jobType"`
	PointData         string   `json:"pointData"`
	JobTime           string   `json:"jobTime"`
	JobPicture        string   `json:"jobPicture"`
	JobAttribute      string   `json:"jobAttribute"`
}
