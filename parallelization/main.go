package main

import (
	"context"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
	"golang.org/x/sync/errgroup"
)

func main() {
	client := openai.NewClient(
		option.WithAPIKey(os.Getenv("API_KEY")),
		option.WithBaseURL(os.Getenv("BASE_URL")),
	)

	codeSnippet := `
func getData(id string) {
    db.Query("SELECT * FROM users WHERE id = " + id)
}
`
	// 1. 定义不同的并行任务
	tasks := []struct {
		name   string
		prompt string
	}{
		{"安全专家", "分析以下代码是否存在 SQL 注入等安全漏洞："},
		{"性能专家", "分析以下代码是否存在性能瓶颈或资源浪费："},
		{"规范专家", "分析以下代码是否符合 Go 语言的代码规范和最佳实践："},
	}

	// 2. 使用 errgroup 并行执行
	g := errgroup.Group{}
	results := make(map[string]string)
	var mu sync.Mutex

	start := time.Now()
	for _, t := range tasks {
		task := t // 闭包变量捕获
		g.Go(func() error {
			fmt.Printf("[%s] 开始分析...\n", task.name)

			resp := callLLM(&client, task.prompt+"\n"+codeSnippet)

			// 安全地存入结果集
			mu.Lock()
			results[task.name] = resp
			mu.Unlock()

			fmt.Printf("[%s] 分析完成。\n", task.name)
			return nil
		})
	}

	// 3. 等待所有并行任务结束
	if err := g.Wait(); err != nil {
		fmt.Printf("并行任务出错: %v\n", err)
		return
	}

	// 4. 聚合阶段 (Aggregation)
	fmt.Printf("\n--- 所有分析已完成，耗时: %v ---\n", time.Since(start))

	finalReport := callLLM(&client, fmt.Sprintf(`
你是一个高级技术主管。
请根据以下三份专家的并行分析报告，汇总成一份最终的、简洁的代码审查意见：

%v
`, results))

	fmt.Println("\n最终聚合报告:\n", finalReport)
}

func callLLM(client *openai.Client, prompt string) string {
	params := openai.ChatCompletionNewParams{
		Model: "deepseek-chat",
		Messages: []openai.ChatCompletionMessageParamUnion{
			openai.UserMessage(prompt),
		},
	}

	resp, err := client.Chat.Completions.New(context.Background(), params)
	if err != nil {
		panic(err)
	}
	return resp.Choices[0].Message.Content
}
