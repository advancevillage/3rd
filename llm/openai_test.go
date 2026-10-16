package llm_test

import (
	"context"
	"os"
	"strings"
	"testing"

	"github.com/advancevillage/3rd/llm"
	"github.com/advancevillage/3rd/logx"
	"github.com/advancevillage/3rd/mathx"
	"github.com/advancevillage/3rd/x"
	"github.com/stretchr/testify/assert"
)

var _ llm.StreamHandler = &testStreamHandler{}

type testStreamHandler struct {
	t    *testing.T
	s    int
	e    int
	id   string // OnStart 捕获的 response id
	text strings.Builder
}

func (t *testStreamHandler) OnStart(ctx context.Context, opts ...x.Option) {
	kv := x.NewBuilder(opts...).Build()
	id, _ := kv[llm.MetaResponseID].(string)
	thinking, _ := kv[llm.MetaThinking].(bool)
	t.id = id
	t.t.Logf("stream started: id=%s thinking=%v", id, thinking)
	t.s += 1
}

func (t *testStreamHandler) OnEnd(ctx context.Context, opts ...x.Option) {
	kv := x.NewBuilder(opts...).Build()
	input, _ := kv[llm.MetaInputTokens].(int64)
	output, _ := kv[llm.MetaOutputTokens].(int64)
	total, _ := kv[llm.MetaTotalTokens].(int64)
	cached, _ := kv[llm.MetaCachedTokens].(int64)
	t.e += 1
	t.t.Logf("stream ended: input=%d output=%d total=%d cached=%d",
		input, output, total, cached)
}

func (t *testStreamHandler) OnChunk(ctx context.Context, chunk string) {
	t.text.WriteString(chunk)
	t.t.Log(chunk)
}

func Test_openai(t *testing.T) {
	if os.Getenv("DOUBAO_SK") == "" {
		t.Skip("DOUBAO_SK is required for the live provider test")
	}
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
				llm.WithSystemMessage("你是一名平面设计专家，突出创意，与众不同。同时兼顾生产制造的限制。"),
				llm.WithUserMessage(`
用户输入： 帮忙生成Q版麻花平面图，1024x1024像素，麦穗形状、纤细，垂直，柔柔的
提取意图： 从用户输入中提取 {{风格类型}} {{尺寸}}
有2种风格模版：(根据【用户输入】匹配相似度，最终输出1个完整的模版)
- 经典
  目标主题：生成一张创意高清背景图
  设计风格：{{风格类型}}
  视觉元素：(优先级由高到低) (模型需要准守布局结构描述)
    P1: 全局背景色是 随机
    P2: 上下2层布局，上部分高度为图片高度的1/3，下部分高度为图片高度的2/3
    P3: 上半部分图案数量不超过10个
    P4: 下半部分无图案
    P5: 150x150像素<=图案大小<=450x450像素 
    P6: 图案方向随机
    P7: 卡通图案
    P8: 图案间间距较大空白间隙
    P9: 主题是....(模型帮忙追加)
    (模型继续增加...)(不少于15条)
  技术参数:
    - 分辨率:  2K    
    - 比例:   3:4   
    - 格式:   PNG   
  负面提示: -无纹理, -无布料效果, -无阴影, -无噪点, -无渐变
-- 麻花
目标主题: 生成一张创意高清背景图
设计风格: {{风格类型}}
视觉元素：(优先级由高到低) (模型需要准守布局结构描述)
  P1: 全局背景色是 随机纯色
  P2: 全局一体化布局
  P3: 550x550像素<=图案大小<=1024x1024像素 
  P4: 图案方向从上到下
  P5: 图案间有间隙，相对稀疏
  P6: 图案主题是 线性创意麻花条纹
  P7: ......(模型帮忙追加)
  (模型继续增加...)(不少于15条)
技术参数:
  - 分辨率:  2K    
  - 比例:   {{尺寸}}  
  - 格式:   PNG   
负面提示: -无纹理, -无布料效果, -无阴影, -无噪点, -无渐变

输出最终模版：（800字以内）(纯文本格式)
模版类型：...(模型帮忙补全且一定存在) 
目标主题：...(模型帮忙补全)
设计风格：...
视觉元素：(优先级由高到低):
  P1: ....
  ........
技术参数:
  ......
负面提示: -无纹理, -无布料效果, -无阴影, -无噪点, -无渐变
				`),
			},
		},
		"case-response": {
			mode: llm.ModeResponse,
			msg: []llm.Message{
				llm.WithSystemMessage("你是一名平面设计专家，突出创意，与众不同。同时兼顾生产制造的限制。"),
				llm.WithUserMessage(`
用户输入： 帮忙生成Q版麻花平面图，1024x1024像素，麦穗形状、纤细，垂直，柔柔的
提取意图： 从用户输入中提取 {{风格类型}} {{尺寸}}
有2种风格模版：(根据【用户输入】匹配相似度，最终输出1个完整的模版)
- 经典
  目标主题：生成一张创意高清背景图
  设计风格：{{风格类型}}
  视觉元素：(优先级由高到低) (模型需要准守布局结构描述)
    P1: 全局背景色是 随机
    P2: 上下2层布局，上部分高度为图片高度的1/3，下部分高度为图片高度的2/3
    P3: 上半部分图案数量不超过10个
    P4: 下半部分无图案
    P5: 150x150像素<=图案大小<=450x450像素 
    P6: 图案方向随机
    P7: 卡通图案
    P8: 图案间间距较大空白间隙
    P9: 主题是....(模型帮忙追加)
    (模型继续增加...)(不少于15条)
  技术参数:
    - 分辨率:  2K    
    - 比例:   3:4   
    - 格式:   PNG   
  负面提示: -无纹理, -无布料效果, -无阴影, -无噪点, -无渐变
-- 麻花
目标主题: 生成一张创意高清背景图
设计风格: {{风格类型}}
视觉元素：(优先级由高到低) (模型需要准守布局结构描述)
  P1: 全局背景色是 随机纯色
  P2: 全局一体化布局
  P3: 550x550像素<=图案大小<=1024x1024像素 
  P4: 图案方向从上到下
  P5: 图案间有间隙，相对稀疏
  P6: 图案主题是 线性创意麻花条纹
  P7: ......(模型帮忙追加)
  (模型继续增加...)(不少于15条)
技术参数:
  - 分辨率:  2K    
  - 比例:   {{尺寸}}  
  - 格式:   PNG   
负面提示: -无纹理, -无布料效果, -无阴影, -无噪点, -无渐变

输出最终模版：（800字以内）(纯文本格式)
模版类型：...(模型帮忙补全且一定存在) 
目标主题：...(模型帮忙补全)
设计风格：...
视觉元素：(优先级由高到低):
  P1: ....
  ........
技术参数:
  ......
负面提示: -无纹理, -无布料效果, -无阴影, -无噪点, -无渐变
				`),
			},
		},
	}

	for n, v := range data {
		f := func(t *testing.T) {
			c, err := llm.NewBaseGPT(ctx, logger,
				llm.WithSecret(os.Getenv("DOUBAO_SK")),
				llm.WithMode(v.mode),
			)
			assert.Nil(t, err)

			h := &testStreamHandler{t: t}
			hh := llm.NewBufferStreamHandler(ctx, logger, h)
			err = c.Completion(ctx, hh, v.msg)
			assert.Nil(t, err)
		}
		t.Run(n, f)
	}
}

// Test_openai_options 验证联网调用以及携带完整历史的多轮请求。
func Test_openai_options(t *testing.T) {
	if os.Getenv("DOUBAO_SK") == "" {
		t.Skip("DOUBAO_SK is required for the live provider test")
	}
	ctx := context.WithValue(context.TODO(), logx.TraceId, mathx.UUID())
	logger, err := logx.NewLogger("debug")
	assert.Nil(t, err)

	c, err := llm.NewBaseGPT(ctx, logger,
		llm.WithSecret(os.Getenv("DOUBAO_SK")),
		llm.WithModel("deepseek-v4-pro-260425"),
	)
	assert.Nil(t, err)

	// 1. 首轮: 联网查询，response id 仅供追踪。
	h1 := &testStreamHandler{t: t}
	hh1 := llm.NewBufferStreamHandler(ctx, logger, h1)
	err = c.Completion(ctx, hh1, []llm.Message{
		llm.WithSystemMessage("你是一名资料调研助手, 回答简洁。"),
		llm.WithUserMessage("帮我查明天武汉天气如何"),
	},
		llm.WithWebSearch(),
	)
	assert.Nil(t, err)
	assert.NotEmpty(t, h1.id)

	// 2. 次轮: 显式发送历史，不依赖供应商存储的 response id。
	h2 := &testStreamHandler{t: t}
	hh2 := llm.NewBufferStreamHandler(ctx, logger, h2)
	err = c.Completion(ctx, hh2, []llm.Message{
		llm.WithSystemMessage("你是一名资料调研助手, 回答简洁。"),
		llm.WithUserMessage("帮我查明天武汉天气如何"),
		llm.WithAssistantMessage(h1.text.String()),
		llm.WithUserMessage("帮我查洪山区天气如何"),
	},
		llm.WithWebSearch(),
	)
	assert.Nil(t, err)
}
