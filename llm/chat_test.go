package llm_test

import (
	"context"
	"os"
	"testing"

	"github.com/advancevillage/3rd/llm"
	"github.com/advancevillage/3rd/logx"
	"github.com/advancevillage/3rd/mathx"
	"github.com/stretchr/testify/assert"
)

var intentSystem = `
你是意图分类助手,默认 chat.text
【cardId 枚举】
- chat.text: 一句话兜底、短应答、普通聊天
- chat.titled: 有明确主题的解释、介绍、定义
- chat.sectioned: 并列要点、分步骤、方法清单
- chat.table: 对比、区别、统计、表格化信息
- chat.cause: 为什么、原因、导致、因果链
- english.show: 英文单词、词义、发音、例句
- english.flashcard: 英文字母认读、大写小写、ABC
- english.story: 英文绘本、英文短故事
- chinese.hanzi: 汉字认读、拼音、笔顺、组词
- chinese.poem: 古诗、唐诗、诗句背诵与讲解
- chinese.story: 中文绘本、短故事、听故事
- math.show: 数学题目展示与口算引导
- math.formula: 数学/物理/生物/化学公式、定理、定律、规律、方程
- math.equation.horizontal: 横式算式加减乘法
- math.equation.vertical: 竖式算式加减乘法
- math.equation.division: 除法算式除法
【输出】
- {"cardId": "{cardId}", "keyword": "{keyword}"}
	`

func Test_chat_chat(t *testing.T) {
	ctx := context.WithValue(context.TODO(), logx.TraceId, mathx.UUID())
	logger, err := logx.NewLogger("debug")
	assert.Nil(t, err)

	data := map[string]struct {
		msg  []llm.Message
		mode string
	}{
		"case-chat": {
			mode: llm.ModeChat,
			msg: []llm.Message{
				llm.WithSystemMessage(intentSystem),
				llm.WithUserMessage("苹果怎么写"),
			},
		},
		"case-chat-2": {
			mode: llm.ModeChat,
			msg: []llm.Message{
				llm.WithSystemMessage(intentSystem),
				llm.WithUserMessage("背诵一下静夜思"),
			},
		},
		"case-chat-3": {
			mode: llm.ModeChat,
			msg: []llm.Message{
				llm.WithSystemMessage(intentSystem),
				llm.WithUserMessage("x^2-3x+1=0"),
			},
		},
		"case-chat-4": {
			mode: llm.ModeChat,
			msg: []llm.Message{
				llm.WithSystemMessage(intentSystem),
				llm.WithUserMessage("光合作用表达式"),
			},
		},
	}

	c, err := llm.NewBaseChat(ctx, logger,
		llm.WithSecret(os.Getenv("DOUBAO_SK")),
		llm.WithMode(llm.ModeChat),
	)
	assert.Nil(t, err)

	for n, v := range data {
		f := func(t *testing.T) {
			reply, err := c.Chat(ctx, v.msg)
			assert.Nil(t, err)
			assert.NotEmpty(t, reply)
			t.Logf("reply: %s", reply)
		}
		t.Run(n, f)
	}
}

func Test_chat_response(t *testing.T) {
	ctx := context.WithValue(context.TODO(), logx.TraceId, mathx.UUID())
	logger, err := logx.NewLogger("debug")
	assert.Nil(t, err)

	data := map[string]struct {
		msg  []llm.Message
		mode string
	}{
		"case-chat": {
			msg: []llm.Message{
				llm.WithSystemMessage(intentSystem),
				llm.WithUserMessage("苹果怎么写"),
			},
		},
		"case-chat-2": {
			msg: []llm.Message{
				llm.WithSystemMessage(intentSystem),
				llm.WithUserMessage("背诵一下静夜思"),
			},
		},
		"case-chat-3": {
			msg: []llm.Message{
				llm.WithSystemMessage(intentSystem),
				llm.WithUserMessage("x^2-3x+1=0"),
			},
		},
		"case-chat-4": {
			msg: []llm.Message{
				llm.WithSystemMessage(intentSystem),
				llm.WithUserMessage("光合作用表达式"),
			},
		},
	}

	c, err := llm.NewBaseChat(ctx, logger,
		llm.WithSecret(os.Getenv("DOUBAO_SK")),
		llm.WithMode(llm.ModeResponse),
	)
	assert.Nil(t, err)

	for n, v := range data {
		f := func(t *testing.T) {
			reply, err := c.Chat(ctx, v.msg)
			assert.Nil(t, err)
			assert.NotEmpty(t, reply)
			t.Logf("reply: %s", reply)
		}
		t.Run(n, f)
	}
}
