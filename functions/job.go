package functions

import (
	"encoding/json"
	"fmt"

	"github.com/go-resty/resty/v2"
)

func GetJobList(arguments map[string]string) string {
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
	Data    []struct {
		ID                string   `json:"id"`
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
		JobAttribute      string   `json:"jobAttribute"`
	} `json:"data"`
}

func GetUserInfo(arguments map[string]string) string {
	var result Result
	resp, _ := resty.New().R().
		SetResult(&result).
		SetQueryParams(arguments).
		Get("https://crm.aifusheng.com/api/user/info")
	body := resp.String()
	fmt.Println(body)
	return body
}

func GetApplyList(arguments map[string]string) string {
	var result Result
	resp, _ := resty.New().R().
		SetResult(&result).
		SetQueryParams(arguments).
		Get("https://crm.aifusheng.com/api/apply/list")
	body := resp.String()
	fmt.Println(body)
	return body
}
