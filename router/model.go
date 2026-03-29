package main

type RouterResponse struct {
	Category string `json:"category"` // 取值范围: technical, billing, chitchat
	Reason   string `json:"reason"`   // 路由理由
}
