package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"net/http"
)

type Detail struct {
	InstanceId string `json:"instance-id"`
	State      string `json:"state"`
}

func handler(ctx context.Context, event events.CloudWatchEvent) (string, error) {
	// 解析 CloudWatch 事件中的 EC2 实例状态更改事件
	var lambdaDetail Detail
	err := json.Unmarshal(event.Detail, &lambdaDetail)
	if err != nil {
		fmt.Println()
		return "fail", fmt.Errorf("Error decoding JSON:", err)
	}
	// 向某个 URL 发送请求
	if err = SendHTTPRequest("http://57.180.139.206/logrus/insert", event); err != nil {
		return "SendHTTPRequest err", err
	}
	var body string
	body, err = GetIP(ctx, lambdaDetail.InstanceId)
	if err != nil {
		body = err.Error()
	}
	// 向某个 URL 发送请求
	if err = SendHTTPRequest("http://57.180.139.206/logrus/insert", body); err != nil {
		return "SendHTTPRequest err", err
	}
	// 在这里执行您的自定义逻辑，比如发送通知等
	return "Function executed successfully!", nil
}

func SendHTTPRequest(url string, payload interface{}) error {
	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonPayload))
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("HTTP Request failed with status code: %d", resp.StatusCode)
	}
	return nil
}

func GetIP(ctx context.Context, instanceID string) (ip string, err error) {
	// 替换为你的 AWS 凭证和区域
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		fmt.Println("Error loading AWS config:", err)
		return
	}
	// 创建 EC2 服务客户端
	svc := ec2.NewFromConfig(cfg)
	// 创建 DescribeInstancesInput 结构体
	// 调用 DescribeInstances 方法获取实例信息
	result, err := svc.DescribeInstances(context.TODO(), &ec2.DescribeInstancesInput{
		InstanceIds: []string{
			instanceID,
		},
	})
	if err != nil {
		fmt.Println("Error describing instances:", err)
		return
	}
	// 提取实例的 IP 地址
	for _, reservation := range result.Reservations {
		for _, instance := range reservation.Instances {
			ip = *instance.PublicIpAddress
		}
	}
	return
}

func main() {
	lambda.Start(handler)
}
