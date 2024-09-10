package dailysign

import (
	"strconv"
	"time"

	"github.com/sirupsen/logrus"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"
)

func (sql *signdb) queryConfig(ctx *zero.Ctx, module string) (err error) {
	msg, _, err := signsql.queryConfigCommon(ctx, module)
	if err != nil {
		return
	}
	msg = append(msg, message.Text("————————"))
	ctx.Send(msg)
	return
}

func (sql *signdb) updateConfig(ctx *zero.Ctx, module string) (err error) {
	msg, configDatas, err := signsql.queryConfigCommon(ctx, module)
	if err != nil {
		return
	}
	msg = append(msg, message.Text("————————\n输入对应序号修改,或回复“取消”取消"))
	ctx.Send(msg)
	recv, cancel := zero.NewFutureEvent("message", 999, false, zero.OnlyGroup, zero.RegexRule(`^(取消|\d+)$`), zero.CheckGroup(ctx.Event.GroupID), zero.CheckUser(ctx.Event.UserID)).Repeat()
	defer cancel()
	check := false
	over := time.NewTimer(30 * time.Second)
	index := 0
	for {
		select {
		case <-over.C:
			ctx.Send(
				message.ReplyWithMessage(ctx.Event.MessageID,
					message.Text("修改配置超时"),
				),
			)
			return
		case m := <-recv:
			cmd := m.Event.Message.String()
			if cmd == "取消" {
				ctx.Send(
					message.ReplyWithMessage(ctx.Event.MessageID,
						message.Text("已取消修改配置"),
					),
				)
				return
			}
			num, _ := strconv.Atoi(cmd)
			if num > len(configDatas)-1 {
				ctx.Send(
					message.ReplyWithMessage(ctx.Event.MessageID,
						message.Text("配置序号不合法"),
					),
				)
				continue
			}
			index = num
			check = true
		}
		if check {
			break
		}
	}
	configData := configDatas[index]
	// 判断配置是否存在
	if !sql.db.CanFind("CONFIG", "where KEY = '"+configData.KEY.String+"' AND MODULE = '"+module+"'") {
		ctx.Send(
			message.ReplyWithMessage(ctx.Event.MessageID,
				message.Text("配置不存在！"),
			),
		)
		return
	}
	ctx.Send(
		message.ReplyWithMessage(ctx.Event.MessageID,
			message.Text("模块："+module+"\nKEY："+configData.KEY.String+"\n————————请输入需要修改的值，或输入“取消”取消修改"),
		),
	)
	recv, cancel = zero.NewFutureEvent("message", 999, false, zero.OnlyGroup, zero.RegexRule(`^(.*)$`), zero.CheckGroup(ctx.Event.GroupID), zero.CheckUser(ctx.Event.UserID)).Repeat()
	defer cancel()
	check = false
	configValue := ""
	over = time.NewTimer(30 * time.Second)
	for {
		select {
		case <-over.C:
			ctx.Send(
				message.ReplyWithMessage(ctx.Event.MessageID,
					message.Text("修改配置超时"),
				),
			)
			return
		case m := <-recv:
			configValue = m.Event.Message.String()
			if configValue == "取消" {
				ctx.Send(
					message.ReplyWithMessage(ctx.Event.MessageID,
						message.Text("已取消修改配置"),
					),
				)
				return
			}
			check = true
		}
		if check {
			break
		}
	}
	configData.VALUE.String = configValue
	sql.db.Insert("CONFIG", &configData)
	ctx.Send(
		message.ReplyWithMessage(ctx.Event.MessageID,
			message.Text("修改配置成功"),
		),
	)
	return
}

func (sql *signdb) queryConfigCommon(ctx *zero.Ctx, module string) (msg message.Message, configDatas []config, err error) {
	sql.Lock()
	defer sql.Unlock()
	err = sql.db.Create("CONFIG", &config{})
	if err != nil {
		logrus.Errorln("[ERROR] queryConfigCommon err: ", err)
		ctx.Send(
			message.ReplyWithMessage(ctx.Event.MessageID,
				message.Text("初始化配置文件失败！"),
			),
		)
		return
	}
	configData := config{}
	err = sql.db.FindFor("CONFIG", &configData, "where MODULE = '"+module+"'", func() error {
		if configData.ID != 0 {
			configDatas = append(configDatas, configData)
		}
		return nil
	})
	if err != nil {
		logrus.Errorln("[ERROR] queryConfigCommon err: ", err)
		ctx.Send(
			message.ReplyWithMessage(ctx.Event.MessageID,
				message.Text("查询配置失败！"),
			),
		)
		return
	}
	if len(configDatas) < 1 {
		ctx.Send(
			message.ReplyWithMessage(ctx.Event.MessageID,
				message.Text("查询无结果"),
			),
		)
		err = errNoResult
		return
	}
	msg = make(message.Message, 0, 3+len(configDatas))
	msg = append(msg, message.Reply(ctx.Event.MessageID), message.Text("找到以下配置:\n"))
	for i, info := range configDatas {
		index := strconv.Itoa(i)
		// 仅保留值五位
		value := info.VALUE.String
		if len(info.VALUE.String) > 5 {
			value = info.VALUE.String[0:5] + "****"
		}
		msg = append(msg, message.Text("["+index+"] "+info.KEY.String+" "+value+"\n"))
	}
	return
}
