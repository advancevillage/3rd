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
	RequestId   string         `json:"request_id"`
	Object      string         `json:"object"`
	CreatedAt   int64          `json:"created_at"`
	CompletedAt int64          `json:"completed_at"`
	Status      string         `json:"status"`
	Data        []txQueryData  `json:"data"`
	Usage       txQueryUsage   `json:"usage"`
}
