package tti

type txSubmitResp struct {
	Id        string `json:"id"`
	RequestId string `json:"request_id"`
	Object    string `json:"object"`
	CreatedAt int64  `json:"created_at"`
	Status    string `json:"status"`
}

type txQueryData struct {
	Url           string `json:"url"`
	RevisedPrompt string `json:"revised_prompt"`
}

type txQueryUsage struct {
	Credits  int `json:"credits"`
	Duration int `json:"duration"`
}

type txQueryResp struct {
	RequestId   string        `json:"request_id"`
	Object      string        `json:"object"`
	CreatedAt   int64         `json:"created_at"`
	CompletedAt int64         `json:"completed_at"`
	Status      string        `json:"status"`
	Data        []txQueryData `json:"data"`
	Usage       txQueryUsage  `json:"usage"`
}

type bdSubmitResp struct {
	Model string         `json:"model"`
	Data  []*bdQueryResp `json:"data"`
	Usage *bdUsageResp   `json:"usage"`
}

type bdQueryResp struct {
	Url  string `json:"url"`
	Size string `json:"size"`
}

type bdUsageResp struct {
	GeneratedImages int `json:"generated_images"`
	OutputTokens    int `json:"output_tokens"`
	TotalTokens     int `json:"total_tokens"`
}
