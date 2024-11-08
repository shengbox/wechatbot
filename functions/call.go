package functions

import (
	"bytes"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/sashabaranov/go-openai"
)

var callMap = map[string]FunctionCall{}

func init() {
	data, err := os.ReadFile("functions.json")
	if err != nil {
		log.Fatal("Error reading file:", err)
	}
	var calls []FunctionCall
	json.Unmarshal(data, &calls)
	for _, it := range calls {
		callMap[it.Name] = it
	}
}

func Call(functionCall *openai.FunctionCall) (string, error) {
	log.Println(functionCall.Name, functionCall.Arguments)

	call := callMap[functionCall.Name]
	var arguments map[string]string
	json.Unmarshal([]byte(functionCall.Arguments), &arguments)
	req, _ := http.NewRequest(call.Method, call.API, bytes.NewReader([]byte(functionCall.Arguments)))
	if strings.ToUpper(call.Method) == "GET" {
		q := req.URL.Query()
		for k, v := range arguments {
			if v != "" && k != "" {
				q.Add(k, v)
			}
		}
		req.URL.RawQuery = q.Encode()
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Println(req.URL.String(), err)
		return "", err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	return string(body), err
}

type FunctionCall struct {
	Name   string `json:"name"`
	API    string `json:"api"`
	Method string `json:"method"`
}
