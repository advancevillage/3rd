package llm_test

import (
	"context"
	"os"
	"slices"
	"testing"

	"github.com/advancevillage/3rd/llm"
	"github.com/advancevillage/3rd/logx"
	"github.com/advancevillage/3rd/mathx"
	"github.com/advancevillage/3rd/x"
	"github.com/stretchr/testify/assert"
)

func Test_completion(t *testing.T) {
	ctx := context.WithValue(context.TODO(), logx.TraceId, mathx.UUID())
	logger, err := logx.NewLogger("debug")
	assert.Nil(t, err)

	cli, err := llm.NewChatGPT(ctx, logger, llm.WitChatGPTSecret(os.Getenv("CHATGPT_KEY")))
	assert.Nil(t, err)

	type ExpectPrompt struct {
		Prompt string `json:"prompt"`
		Style  string `json:"style"`
	}

	var data = map[string]struct {
		role   string
		query  string
		schema x.Builder
		expect any
	}{
		"case1": {
			role:  "你是一名提示词专家，请从用户输入中提取意图和关键特征，并给出适合tripo的提示词, 请使用英语输出。",
			query: "请帮我生成一个可爱的卡通缅因猫",
			schema: x.NewBuilder(
				x.WithKV("type", "object"),
				x.WithKV("properties", x.NewBuilder(
					x.WithKV("prompt", x.NewBuilder(
						x.WithKV("type", "string"),
						x.WithKV("description", "英文输出, 最多1000字符"),
					).Build()),
					x.WithKV("style", x.NewBuilder(
						x.WithKV("type", "string"),
						x.WithKV("enum", []string{
							"person:person2cartoon",
							"object:clay",
							"object:steampunk",
							"animal:venom",
							"object:barbie",
							"object:christmas",
							"gold",
							"ancient_bronze",
						}),
					).Build()),
				).Build()),
				x.WithKV("required", []string{
					"prompt",
					"style",
				}),
				x.WithKV("additionalProperties", false),
			),
			expect: &ExpectPrompt{},
		},
	}
	for n, v := range data {
		f := func(t *testing.T) {
			err = cli.Completion(ctx, v.role, v.query, v.schema, v.expect)
			assert.Nil(t, err)
			t.Log(v.expect)
			assert.Equal(t, true, len(v.expect.(*ExpectPrompt).Prompt) > 0)
			assert.Equal(t, true, slices.Contains([]string{
				"person:person2cartoon",
				"object:clay",
				"object:steampunk",
				"animal:venom",
				"object:barbie",
				"object:christmas",
				"gold",
				"ancient_bronze",
			}, v.expect.(*ExpectPrompt).Style))
		}
		t.Run(n, f)
	}
}
