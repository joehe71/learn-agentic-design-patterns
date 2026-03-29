package main

type Plan struct {
	Tasks []string `json:"tasks"`
}

type TaskResult struct {
	Task   string `json:"task"`
	Result any    `json:"result"`
}

type ExecutionResult struct {
	Results []TaskResult `json:"results"`
}
