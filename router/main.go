package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
)

func main() {
	client := openai.NewClient(
		option.WithAPIKey(os.Getenv("API_KEY")),
		option.WithBaseURL(os.Getenv("BASE_URL")),
	)
	ctx := context.Background()

	// 模拟三个不同的用户输入
	inputs := []string{
		//"我的 Kubernetes Pod 一直处于 Pending 状态怎么排查？",
		//"我上个月的账单多扣了 5 块钱，帮我查查",
		"今天天气真不错啊！",
	}

	for _, input := range inputs {
		fmt.Printf("\n--- 用户输入: %s ---\n", input)

		// Step 1: 路由判断 (The Router)
		route, err := getRoute(ctx, &client, input)
		if err != nil {
			panic(err)
		}
		fmt.Printf("[路由结果]: 分发至 -> %s (原因: %s)\n", route.Category, route.Reason)

		// Step 2: 根据路由结果进入不同分支
		switch route.Category {
		case "technical":
			handleTechnical(ctx, &client, input)
		case "billing":
			handleBilling(ctx, &client, input)
		case "chitchat":
			handleChitchat(input)
		default:
			fmt.Println("无法识别的任务类型")
		}
	}
}

// getRoute 充当路由器的角色
func getRoute(ctx context.Context, client *openai.Client, input string) (*RouterResponse, error) {
	prompt := fmt.Sprintf(
		`
你是一个任务分拣器。请根据用户输入，将其归类为以下三种类型之一：
- technical: 涉及编程、服务器、架构等技术问题。
- billing: 涉及价格、订单、退款、发票等财务问题。
- chitchat: 简单的问候、天气或无实质内容的闲聊。

要求：严格输出 JSON 格式。
用户输入: "%s"`,
		input,
	)

	answer, err := callLLM(ctx, client, prompt, true)
	if err != nil {
		return nil, err
	}
	response := new(RouterResponse)
	if err := json.Unmarshal([]byte(answer), response); err != nil {
		return nil, err
	}
	return response, nil
}

// --- 以下是不同的业务处理分支 ---

func handleTechnical(ctx context.Context, client *openai.Client, input string) {
	fmt.Println("[执行逻辑]: 正在调用技术专家 Prompt 和 K8s 知识库...")
	answer, err := callLLM(ctx, client, "你是一位资深架构师，请回答："+input, false)
	if err != nil {
		panic(err)
	}
	fmt.Println("解答:", answer)
}

func handleBilling(ctx context.Context, client *openai.Client, input string) {
	fmt.Println("[执行逻辑]: 正在接入财务系统 API 并准备退款策略...")
	// 这里可以写具体的账单处理逻辑
	fmt.Println("解答: 您的账单请求已收到，财务专员正在审核。")
}

func handleChitchat(input string) {
	fmt.Println("[执行逻辑]: 无需调用昂贵模型，直接返回预设话术或轻量回复。")
	fmt.Println("解答: 确实呢！心情也跟着变好了。")
}

func callLLM(ctx context.Context, client *openai.Client, prompt string, isJSON bool) (string, error) {
	params := openai.ChatCompletionNewParams{
		Model: "deepseek-chat",
		Messages: []openai.ChatCompletionMessageParamUnion{
			openai.UserMessage(prompt),
		},
	}

	if isJSON {
		params.ResponseFormat = openai.ChatCompletionNewParamsResponseFormatUnion{
			OfJSONObject: &openai.ResponseFormatJSONObjectParam{},
		}
	}

	resp, err := client.Chat.Completions.New(ctx, params)
	if err != nil {
		return "", err
	}
	return resp.Choices[0].Message.Content, nil
}
