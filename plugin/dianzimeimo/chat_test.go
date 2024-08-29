package dianzimeimo

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/agiledragon/gomonkey"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"
)

func TestStartChat(t *testing.T) {
	var ctx *zero.Ctx
	ctxSendMock := gomonkey.ApplyMethod(reflect.TypeOf(ctx), "Send", func(*zero.Ctx, interface{}) message.MessageID {
		return message.MessageID{}
	})
	defer ctxSendMock.Reset()
	chatdata.startChat(ctx)
}

func TestChat(t *testing.T) {
	ctx := &zero.Ctx{
		Event: &zero.Event{
			UserID: 408029164,
		},
		State: zero.State{
			"regex_matched": []string{"chat", "这里是哪里 你是谁"},
		},
	}
	getChatMock := gomonkey.ApplyFunc(getUserChatInfo, func(*zero.Ctx, *chatdb) []Message {
		return []Message{baseMessageMap["startMessage"]}
	})
	defer getChatMock.Reset()
	chatdata.userChat(ctx)
}

func TestReadResponseByLines(t *testing.T) {
	text := `data:  {"choices":[{"delta":{"content":"*"}}]}

data:  {"choices":[{"delta":{"content":"派"}}]}

data:  {"choices":[{"delta":{"content":"蒙"}}]}

data:  {"choices":[{"delta":{"content":"脸上"}}]}

data:  {"choices":[{"delta":{"content":"浮现"}}]}

data:  {"choices":[{"delta":{"content":"出"}}]}

data:  {"choices":[{"delta":{"content":"温柔"}}]}

data:  {"choices":[{"delta":{"content":"而"}}]}

data:  {"choices":[{"delta":{"content":"神秘"}}]}

data:  {"choices":[{"delta":{"content":"的笑容"}}]}

data:  {"choices":[{"delta":{"content":"。"}}]}

data:  {"choices":[{"delta":{"content":"*"}}]}

data:  {"choices":[{"delta":{"content":"“"}}]}

data:  {"choices":[{"delta":{"content":"我是"}}]}

data:  {"choices":[{"delta":{"content":"这片"}}]}

data:  {"choices":[{"delta":{"content":"提"}}]}

data:  {"choices":[{"delta":{"content":"瓦"}}]}

data:  {"choices":[{"delta":{"content":"特"}}]}

data:  {"choices":[{"delta":{"content":"大陆"}}]}

data:  {"choices":[{"delta":{"content":"的"}}]}

data:  {"choices":[{"delta":{"content":"守护"}}]}

data:  {"choices":[{"delta":{"content":"者"}}]}

data:  {"choices":[{"delta":{"content":"，"}}]}

data:  {"choices":[{"delta":{"content":"也是"}}]}

data:  {"choices":[{"delta":{"content":"这片"}}]}

data:  {"choices":[{"delta":{"content":"土地"}}]}

data:  {"choices":[{"delta":{"content":"上的"}}]}

data:  {"choices":[{"delta":{"content":"精灵"}}]}

data:  {"choices":[{"delta":{"content":"之一"}}]}

data:  {"choices":[{"delta":{"content":"。"}}]}

data:  {"choices":[{"delta":{"content":"我"}}]}

data:  {"choices":[{"delta":{"content":"以"}}]}

data:  {"choices":[{"delta":{"content":"森林"}}]}

data:  {"choices":[{"delta":{"content":"和"}}]}

data:  {"choices":[{"delta":{"content":"自然的"}}]}

data:  {"choices":[{"delta":{"content":"形象"}}]}

data:  {"choices":[{"delta":{"content":"存在"}}]}

data:  {"choices":[{"delta":{"content":"，"}}]}

data:  {"choices":[{"delta":{"content":"为"}}]}

data:  {"choices":[{"delta":{"content":"这个"}}]}

data:  {"choices":[{"delta":{"content":"美丽"}}]}

data:  {"choices":[{"delta":{"content":"的大"}}]}

data:  {"choices":[{"delta":{"content":"陆"}}]}

data:  {"choices":[{"delta":{"content":"带来"}}]}

data:  {"choices":[{"delta":{"content":"生机"}}]}

data:  {"choices":[{"delta":{"content":"和"}}]}

data:  {"choices":[{"delta":{"content":"活力"}}]}

data:  {"choices":[{"delta":{"content":"。"}}]}

data:  {"choices":[{"delta":{"content":"”"}}]}

data:  {"choices":[{"delta":{"content":" "}}]}

data:  {"choices":[{"delta":{"content":"她"}}]}

data:  {"choices":[{"delta":{"content":"伸出"}}]}

data:  {"choices":[{"delta":{"content":"修"}}]}

data:  {"choices":[{"delta":{"content":"长的"}}]}

data:  {"choices":[{"delta":{"content":"手指"}}]}

data:  {"choices":[{"delta":{"content":"，"}}]}

data:  {"choices":[{"delta":{"content":"轻轻"}}]}

data:  {"choices":[{"delta":{"content":"拂"}}]}

data:  {"choices":[{"delta":{"content":"过"}}]}

data:  {"choices":[{"delta":{"content":"你的"}}]}

data:  {"choices":[{"delta":{"content":"脸"}}]}

data:  {"choices":[{"delta":{"content":"颊"}}]}

data:  {"choices":[{"delta":{"content":"。"}}]}

data:  {"choices":[{"delta":{"content":"*"}}]}

data:  {"choices":[{"delta":{"content":"“"}}]}

data:  {"choices":[{"delta":{"content":"我在"}}]}

data:  {"choices":[{"delta":{"content":"森林"}}]}

data:  {"choices":[{"delta":{"content":"里"}}]}

data:  {"choices":[{"delta":{"content":"寻找"}}]}

data:  {"choices":[{"delta":{"content":"着你"}}]}

data:  {"choices":[{"delta":{"content":"，"}}]}

data:  {"choices":[{"delta":{"content":"发现"}}]}

data:  {"choices":[{"delta":{"content":"你"}}]}

data:  {"choices":[{"delta":{"content":"受了"}}]}

data:  {"choices":[{"delta":{"content":"重伤"}}]}

data:  {"choices":[{"delta":{"content":"。"}}]}

data:  {"choices":[{"delta":{"content":"看到"}}]}

data:  {"choices":[{"delta":{"content":"你"}}]}

data:  {"choices":[{"delta":{"content":"濒"}}]}

data:  {"choices":[{"delta":{"content":"临"}}]}

data:  {"choices":[{"delta":{"content":"危险"}}]}

data:  {"choices":[{"delta":{"content":"，"}}]}

data:  {"choices":[{"delta":{"content":"我就"}}]}

data:  {"choices":[{"delta":{"content":"用自己的"}}]}

data:  {"choices":[{"delta":{"content":"魔法"}}]}

data:  {"choices":[{"delta":{"content":"来"}}]}

data:  {"choices":[{"delta":{"content":"治疗"}}]}

data:  {"choices":[{"delta":{"content":"你"}}]}

data:  {"choices":[{"delta":{"content":"。"}}]}

data:  {"choices":[{"delta":{"content":"现在"}}]}

data:  {"choices":[{"delta":{"content":"你"}}]}

data:  {"choices":[{"delta":{"content":"正在"}}]}

data:  {"choices":[{"delta":{"content":"恢复"}}]}

data:  {"choices":[{"delta":{"content":"，"}}]}

data:  {"choices":[{"delta":{"content":"但你"}}]}

data:  {"choices":[{"delta":{"content":"还需要"}}]}

data:  {"choices":[{"delta":{"content":"更多"}}]}

data:  {"choices":[{"delta":{"content":"的时间"}}]}

data:  {"choices":[{"delta":{"content":"来"}}]}

data:  {"choices":[{"delta":{"content":"完全"}}]}

data:  {"choices":[{"delta":{"content":"康复"}}]}

data:  {"choices":[{"delta":{"content":"。"}}]}

data:  {"choices":[{"delta":{"content":"在这个"}}]}

data:  {"choices":[{"delta":{"content":"过程中"}}]}

data:  {"choices":[{"delta":{"content":"，"}}]}

data:  {"choices":[{"delta":{"content":"你会"}}]}

data:  {"choices":[{"delta":{"content":"感到"}}]}

data:  {"choices":[{"delta":{"content":"疲惫"}}]}

data:  {"choices":[{"delta":{"content":"，"}}]}

data:  {"choices":[{"delta":{"content":"所以我"}}]}

data:  {"choices":[{"delta":{"content":"让你"}}]}

data:  {"choices":[{"delta":{"content":"在这里"}}]}

data:  {"choices":[{"delta":{"content":"好好"}}]}

data:  {"choices":[{"delta":{"content":"休息"}}]}

data:  {"choices":[{"delta":{"content":"。"}}]}

data:  {"choices":[{"delta":{"content":"我会"}}]}

data:  {"choices":[{"delta":{"content":"一直"}}]}

data:  {"choices":[{"delta":{"content":"陪"}}]}

data:  {"choices":[{"delta":{"content":"在你"}}]}

data:  {"choices":[{"delta":{"content":"身边"}}]}

data:  {"choices":[{"delta":{"content":"，"}}]}

data:  {"choices":[{"delta":{"content":"为你"}}]}

data:  {"choices":[{"delta":{"content":"提供"}}]}

data:  {"choices":[{"delta":{"content":"支持和"}}]}

data:  {"choices":[{"delta":{"content":"安慰"}}]}

data:  {"choices":[{"delta":{"content":"。"}}]}

data:  {"choices":[{"delta":{"content":"”"}}]}

data:  {"choices":[{"delta":{"content":""}}]}

`
	data := []byte(text)
	res, _ := readResponseByLines(data)
	fmt.Println(res)
	translateResponse, _ := translate(res)
	fmt.Println(translateResponse)
}
