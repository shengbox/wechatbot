package functions

import (
	"bytes"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
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

func Call(name string, arguments map[string]string) (string, error) {
	call := callMap[name]

	reqBody, _ := json.Marshal(arguments)
	req, _ := http.NewRequest(call.Method, call.API, bytes.NewReader(reqBody))
	if strings.ToUpper(call.Method) == "GET" {
		q := req.URL.Query()
		for k, v := range arguments {
			q.Add(k, v)
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
