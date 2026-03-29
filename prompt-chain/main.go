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

	userInput := "帮我分析 Kubernetes，并给出适用场景"

	// Step1: 任务拆解
	planJSON := callLLM(ctx, &client, fmt.Sprintf(`
你是一个任务规划器。
请把用户问题拆解成多个可执行步骤。

要求：
1. 输出 JSON
2. 格式如下：
{
  "tasks": ["step1", "step2"]
}

用户输入：
%s
`, userInput))

	fmt.Println("Plan Raw:", planJSON)

	var plan Plan
	mustParseJSON(planJSON, &plan)

	// Step2: 执行每个任务
	var results []TaskResult
	for _, task := range plan.Tasks {
		resultJSON := callLLM(ctx, &client, fmt.Sprintf(`
你是一个执行器。
请完成以下任务，并返回 JSON：

{
  "task": "%s",
  "result": "..."
}
`, task))

		var tr TaskResult
		mustParseJSON(resultJSON, &tr)

		results = append(results, tr)
	}

	execResult := ExecutionResult{Results: results}

	// Step3: 汇总
	final := callLLM(ctx, &client, fmt.Sprintf(`
你是一个总结器。

根据以下 JSON 数据生成最终回答：

%s
`, toJSON(execResult)))

	fmt.Println("Final:\n", final)
}

func callLLM(ctx context.Context, client *openai.Client, prompt string) string {
	resp, err := client.Chat.Completions.New(ctx, openai.ChatCompletionNewParams{
		Model: "deepseek-chat",
		Messages: []openai.ChatCompletionMessageParamUnion{
			openai.UserMessage(prompt),
		},
		ResponseFormat: openai.ChatCompletionNewParamsResponseFormatUnion{
			OfJSONObject: &openai.ResponseFormatJSONObjectParam{},
		},
	})
	if err != nil {
		panic(err)
	}

	return resp.Choices[0].Message.Content
}

func mustParseJSON(s string, v any) {
	err := json.Unmarshal([]byte(s), v)
	if err != nil {
		panic(fmt.Sprintf("JSON parse error: %v\nraw: %s", err, s))
	}
}

func toJSON(v any) string {
	b, _ := json.MarshalIndent(v, "", "  ")
	return string(b)
}
