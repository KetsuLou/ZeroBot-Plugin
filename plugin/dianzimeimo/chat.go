// 电子魅魔
package dianzimeimo

import (
	"bufio"
	"bytes"
	"encoding/json"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/FloatTech/floatbox/binary"
	fcext "github.com/FloatTech/floatbox/ctxext"
	"github.com/FloatTech/floatbox/web"
	sql "github.com/FloatTech/sqlite"
	ctrl "github.com/FloatTech/zbpctrl"
	"github.com/FloatTech/zbputils/control"
	"github.com/FloatTech/zbputils/ctxext"
	"github.com/sirupsen/logrus"
	"github.com/tidwall/gjson"

	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"
)

type chatdb struct {
	db *sql.Sqlite
	sync.RWMutex
}

type userChatInfo struct {
	ID       int64  // 主键
	UserID   int64  // 用户ID
	ChatInfo string // 对话内容
	Date     string // 日期
}

type chatBaseJson struct {
	Messages                   []Message `json:"messages"`
	Model                      string    `json:"model"`
	Temperature                int       `json:"temperature"`
	FrequencyPenalty           int       `json:"frequency_penalty"`
	PresencePenalty            int       `json:"presence_penalty"`
	TopP                       int       `json:"top_p"`
	MaxTokens                  int       `json:"max_tokens"`
	Stream                     bool      `json:"stream"`
	ChatCompletionSource       string    `json:"chat_completion_source"`
	UserName                   string    `json:"user_name"`
	CharName                   string    `json:"char_name"`
	CustomUrl                  string    `json:"custom_url"`
	CustomIncludeBody          string    `json:"custom_include_body"`
	CustomExcludeBody          string    `json:"custom_exclude_body"`
	CustomIncludeHeaders       string    `json:"custom_include_headers"`
	CustomPromptPostProcessing string    `json:"custom_prompt_post_processing"`
	ReverseProxy               string    `json:"reverse_proxy"`
	ProxyPassword              string    `json:"proxy_password"`
}

const chatCompletionsURL = "/chat/completions"

var (
	chatdata = &chatdb{
		db: &sql.Sqlite{},
	}
	engine = control.Register("dianzimeimo", &ctrl.Options[*zero.Ctx]{
		DisableOnDefault: false,
		Brief:            "电子魅魔",
		Help: "电子魅魔\n----------指令----------\n" +
			"- 开始对话\n" +
			"- 重新回复\n" +
			"- 返回上轮对话\n" +
			"- chat [对话内容]",
		PrivateDataFolder: "dianzimeimo",
	}).ApplySingle(ctxext.DefaultSingle)
	getdb = fcext.DoOnceOnSuccess(func(ctx *zero.Ctx) bool {
		chatdata.db.DBPath = engine.DataFolder() + "dianzimeimo.db"
		err := chatdata.db.Open(time.Hour * 24)
		if err != nil {
			ctx.SendChain(message.Text("[ERROR]:", err))
			return false
		}
		return true
	})
)

func init() {
	engine.OnRegex(`^开始对话$`, zero.SuperUserPermission, getdb).SetBlock(true).Limit(ctxext.LimitByGroup).Handle(func(ctx *zero.Ctx) {
		chatdata.startChat(ctx)
	})
	engine.OnRegex(`^重新回复$`, zero.SuperUserPermission, getdb).SetBlock(true).Limit(ctxext.LimitByGroup).Handle(func(ctx *zero.Ctx) {
		chatdata.regenerateChat(ctx)
	})
	engine.OnRegex(`^返回上轮对话$`, zero.SuperUserPermission, getdb).SetBlock(true).Limit(ctxext.LimitByGroup).Handle(func(ctx *zero.Ctx) {
		chatdata.startChat(ctx)
	})
	// 对话内容
	engine.OnRegex(`^chat\s*(.*)$`, zero.SuperUserPermission, getdb).SetBlock(true).Limit(ctxext.LimitByGroup).Handle(func(ctx *zero.Ctx) {
		chatdata.userChat(ctx)
	})
}

func (sql *chatdb) startChat(ctx *zero.Ctx) (err error) {
	err = sql.db.Create("userChat", &userChatInfo{})
	if err != nil {
		logrus.Errorln("[ERROR] create userChat err: ", err)
		ctx.Send(
			message.ReplyWithMessage(ctx.Event.MessageID,
				message.Text("创建用户对话信息表失败！"),
			),
		)
		return
	}
	chatInfos := getUserChatInfo(ctx, sql)
	if len(chatInfos) > 0 {
		ctx.Send(
			message.ReplyWithMessage(ctx.Event.MessageID,
				message.Text("您有正在进行的对话！"),
			),
		)
		lastMessage := chatInfos[len(chatInfos)-1]
		msg, _ := translate(lastMessage.Content)
		ctx.Send(
			message.ReplyWithMessage(ctx.Event.MessageID,
				message.Text(msg),
			),
		)
		return
	}
	// 增加初始对话
	insertUserChatInfo(ctx, paimengMessageMap["startMessage"], sql)
	ctx.Send(
		message.ReplyWithMessage(ctx.Event.MessageID,
			message.Text(paimengMessageMap["startMessage"].Content),
		),
	)
	return
}

func (sql *chatdb) regenerateChat(ctx *zero.Ctx) (err error) {
	chatInfos := getUserChatInfo(ctx, sql)
	if len(chatInfos) == 0 {
		ctx.Send(
			message.ReplyWithMessage(ctx.Event.MessageID,
				message.Text("您没有正在进行的对话，请先输入【开始对话】开始对话"),
			),
		)
		return
	}
	body := jsonToStuct(xiaoshuoaiMap["baseJson"])
	messagesData := append(paimengMessageList, chatInfos[:len(chatInfos)-2]...)
	body.Messages = messagesData
	b, err := json.Marshal(body)
	if err != nil {
		logrus.Errorln("regenerateChat json Marshal error.")
		return err
	}
	// logrus.Infoln(body)
	response, err := aiReply(ctx, b)
	if err != nil {
		return err
	}
	content, _ := readResponseByLines(response)
	msg, _ := translate(content)
	ctx.Send(
		message.ReplyWithMessage(ctx.Event.MessageID,
			message.Text(msg),
		),
	)
	message := Message{}
	message.Role = "assistant"
	message.Content = content
	return insertUserChatInfo(ctx, message, sql)
}

func (sql *chatdb) userChat(ctx *zero.Ctx) (err error) {
	text := ctx.State["regex_matched"].([]string)[1]
	chatInfos := getUserChatInfo(ctx, sql)
	if len(chatInfos) == 0 {
		ctx.Send(
			message.ReplyWithMessage(ctx.Event.MessageID,
				message.Text("您没有正在进行的对话，请先输入【开始对话】开始对话"),
			),
		)
		return
	}
	body := jsonToStuct(xiaoshuoaiMap["baseJson"])
	userMessage := Message{
		Role:    "user",
		Content: text,
	}
	insertUserChatInfo(ctx, userMessage, sql)
	messagesData := append(paimengMessageList, chatInfos...)
	messagesData = append(messagesData, userMessage)
	body.Messages = messagesData
	b, err := json.Marshal(body)
	if err != nil {
		return err
	}
	// logrus.Infoln(body)
	response, err := aiReply(ctx, b)
	if err != nil {
		return
	}

	content, _ := readResponseByLines(response)
	msg, _ := translate(content)
	ctx.Send(
		message.ReplyWithMessage(ctx.Event.MessageID,
			message.Text(msg),
		),
	)
	message := Message{}
	message.Role = "assistant"
	message.Content = content
	return insertUserChatInfo(ctx, message, sql)
}

// 从数据库查询用户对话信息
func getUserChatInfo(ctx *zero.Ctx, sql *chatdb) (chatMessages []Message) {
	sql.Lock()
	defer sql.Unlock()
	userChatInfoList := []userChatInfo{}
	chatInfo := userChatInfo{}
	sql.db.FindFor("userChat", &chatInfo, "WHERE UserID = "+strconv.FormatInt(ctx.Event.UserID, 10)+" ORDER BY Date DESC LIMIT 4", func() error {
		userChatInfoList = append(userChatInfoList, chatInfo)
		return nil
	})
	if len(userChatInfoList) == 0 {
		return chatMessages
	}
	for _, info := range userChatInfoList {
		msg := Message{}
		err := json.Unmarshal([]byte(info.ChatInfo), &msg)
		if err != nil {
			logrus.Errorln("[ERROR] chatInfoToStuct err: ", err)
			msg = Message{}
		}
		chatMessages = append(chatMessages, msg)
	}
	// 逆序
	sort.Slice(chatMessages, func(i, j int) bool {
		return true
	})
	return
}

func insertUserChatInfo(ctx *zero.Ctx, chatInfo Message, sql *chatdb) (err error) {
	sql.Lock()
	defer sql.Unlock()
	jsonData, _ := json.Marshal(chatInfo)
	userChat := userChatInfo{
		ID:       time.Now().Unix(),
		UserID:   ctx.Event.UserID,
		ChatInfo: string(jsonData),
		Date:     time.Now().Format("2006-01-02 15:04:05"),
	}
	return sql.db.Insert("userChat", &userChat)
}

// 发送信息字符串转结构体
func jsonToStuct(value string) (info chatBaseJson) {
	err := json.Unmarshal([]byte(value), &info)
	if err != nil {
		logrus.Errorln("[ERROR] jsonToStuct err: ", err)
		info = chatBaseJson{}
	}
	return
}

func aiReply(ctx *zero.Ctx, data []byte) (response []byte, err error) {
	response, err = web.RequestDataWithHeaders(http.DefaultClient, xiaoshuoaiMap["url"]+chatCompletionsURL, "POST",
		func(r *http.Request) error {
			r.Header.Set("Authorization", xiaoshuoaiMap["Authorization"])
			r.Header.Set("Content-Type", "application/json")
			return nil
		}, bytes.NewReader(data))
	if err != nil {
		logrus.Errorln("userChat chatgpt request error.")
		ctx.SendChain(message.Text("ERROR: ", err))
	}
	return
}

func readResponseByLines(data []byte) (string, error) {
	lines := make([]string, 0)
	reader := bufio.NewReader(bytes.NewReader(data))
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			// 当读取到EOF时，err会是io.EOF，这时可以安全退出循环
			if err.Error() == "EOF" {
				break
			}
			return "", err
		}
		text := strings.Replace(line, "\n", "", -1)
		text = strings.Replace(text, "\t", "", -1)
		text = strings.Replace(text, "data:  ", "", -1)
		if text == "" {
			continue
		}
		lines = append(lines, text)
	}
	textList := []string{}
	for _, line := range lines {
		textList = append(textList, gjson.Get(line, "choices.0.delta.content").String())
	}
	return strings.Join(textList, ""), nil
}

func translate(text string) (msg string, err error) {
	// 初始赋值,翻译失败则返回原文本
	msg = text
	body := Translate{
		Text:       text,
		TargetLang: "ZH",
	}
	b, err := json.Marshal(body)
	if err != nil {
		return
	}
	response, err := web.RequestDataWithHeaders(http.DefaultClient, xiaoshuoaiMap["translateURL"], "POST",
		func(r *http.Request) error {
			r.Header.Set("Content-Type", "application/json")
			return nil
		}, bytes.NewReader(b))
	if err != nil {
		logrus.Errorln("[ERROR] translate err: ", err)
		return
	}
	msg = gjson.Get(binary.BytesToString(response), "data").String()
	return
}

func cleanUserChat(ctx *zero.Ctx, sql *chatdb) {
	userChatInfoList := []userChatInfo{}
	chatInfo := userChatInfo{}
	sql.db.FindFor("userChat", &chatInfo, "WHERE UserID = "+strconv.FormatInt(ctx.Event.UserID, 10)+" ORDER BY Date DESC LIMIT 4", func() error {
		userChatInfoList = append(userChatInfoList, chatInfo)
		return nil
	})
	// 删除所有对话内容
	// sql.db.Del("userChat", "WHERE UserID = "+strconv.FormatInt(ctx.Event.UserID, 10))
	for _, info := range userChatInfoList {
		sql.db.Insert("userChat", &info)
	}
}
