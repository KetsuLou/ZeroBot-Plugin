package dailysign

import (
	"strconv"
	"time"

	"github.com/sirupsen/logrus"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"
)

func (sql *signdb) queryState(ctx *zero.Ctx, module string) (err error) {
	msg, stateDatas, err := signsql.queryStateCommon(ctx, module)
	if err != nil {
		return
	}
	msg = append(msg, message.Text("————————\n输入对应序号查看,或回复“取消”取消"))
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
					message.Text("查看状态超时"),
				),
			)
			return
		case m := <-recv:
			cmd := m.Event.Message.String()
			if cmd == "取消" {
				ctx.Send(
					message.ReplyWithMessage(ctx.Event.MessageID,
						message.Text("已取消查看状态"),
					),
				)
				return
			}
			num, _ := strconv.Atoi(cmd)
			if num > len(stateDatas)-1 {
				ctx.Send(
					message.ReplyWithMessage(ctx.Event.MessageID,
						message.Text("状态序号不合法"),
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
	stateData := stateDatas[index]
	// 判断状态是否存在
	if !sql.db.CanFind("STATE", "where NAME = '"+stateData.NAME+"'") {
		ctx.Send(
			message.ReplyWithMessage(ctx.Event.MessageID,
				message.Text("状态不存在！"),
			),
		)
		return
	}
	ctx.Send(
		message.ReplyWithMessage(ctx.Event.MessageID,
			message.Text("模块: ", module,
				"\n名称: ", stateData.NAME,
				"\n是否启用: ", stateData.ENABLE.String,
				"\n运行状态: ", stateData.STATE.String,
				"\n标题: ", stateData.TITLE.String,
				"\n信息: ", stateData.INFO.String,
				"\n下次运行时间: ", stateData.NEXT.String,
				"\n最后运行时间: ", stateData.DATE.String),
		),
	)
	return
}

func (sql *signdb) updateState(ctx *zero.Ctx, module string) (err error) {
	msg, stateDatas, err := signsql.queryStateCommon(ctx, module)
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
					message.Text("修改状态超时"),
				),
			)
			return
		case m := <-recv:
			cmd := m.Event.Message.String()
			if cmd == "取消" {
				ctx.Send(
					message.ReplyWithMessage(ctx.Event.MessageID,
						message.Text("已取消修改状态"),
					),
				)
				return
			}
			num, _ := strconv.Atoi(cmd)
			if num > len(stateDatas)-1 {
				ctx.Send(
					message.ReplyWithMessage(ctx.Event.MessageID,
						message.Text("状态序号不合法"),
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
	stateData := stateDatas[index]
	// 判断配置是否存在
	if !sql.db.CanFind("STATE", "where NAME = '"+stateData.NAME+"' AND MODULE = '"+module+"'") {
		ctx.Send(
			message.ReplyWithMessage(ctx.Event.MessageID,
				message.Text("状态不存在！"),
			),
		)
		return
	}
	ctx.Send(
		message.ReplyWithMessage(ctx.Event.MessageID,
			message.Text("模块: ", module,
				"\n名称: ", stateData.NAME,
				"\n是否启用: ", stateData.ENABLE.String,
				"\n运行状态: ", stateData.STATE.String,
				"\n标题: ", stateData.TITLE.String,
				"\n信息: ", stateData.INFO.String,
				"\n下次运行时间: ", stateData.NEXT.String,
				"\n最后运行时间: ", stateData.DATE.String,
				"\n————————\n输入修改的启用状态（开启/关闭）,或回复“取消”取消"),
		),
	)
	recv, cancel = zero.NewFutureEvent("message", 999, false, zero.OnlyGroup, zero.RegexRule(`^(开启|关闭|取消)$`), zero.CheckGroup(ctx.Event.GroupID), zero.CheckUser(ctx.Event.UserID)).Repeat()
	defer cancel()
	check = false
	enableValue := ""
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
			enableValue = m.Event.Message.String()
			if enableValue == "取消" {
				ctx.Send(
					message.ReplyWithMessage(ctx.Event.MessageID,
						message.Text("已取消修改状态配置"),
					),
				)
				return
			}
			if enableValue == "开启" {
				enableValue = "True"
			}
			if enableValue == "关闭" {
				enableValue = "False"
			}
			check = true
		}
		if check {
			break
		}
	}
	stateData.ENABLE.String = enableValue
	sql.db.Insert("STATE", &stateData)
	ctx.Send(
		message.ReplyWithMessage(ctx.Event.MessageID,
			message.Text("修改状态配置成功"),
		),
	)
	return
}

func (sql *signdb) queryStateCommon(ctx *zero.Ctx, module string) (msg message.Message, stateDatas []state, err error) {
	sql.Lock()
	defer sql.Unlock()
	err = sql.db.Create("STATE", &state{})
	if err != nil {
		logrus.Errorln("[ERROR] queryStateCommon err: ", err)
		ctx.Send(
			message.ReplyWithMessage(ctx.Event.MessageID,
				message.Text("初始化状态文件失败！"),
			),
		)
		return
	}
	stateData := state{}
	err = sql.db.FindFor("STATE", &stateData, "where MODULE = '"+module+"'", func() error {
		if stateData.NAME != "" {
			stateDatas = append(stateDatas, stateData)
		}
		return nil
	})
	if err != nil {
		logrus.Errorln("[ERROR] queryStateCommon err: ", err)
		ctx.Send(
			message.ReplyWithMessage(ctx.Event.MessageID,
				message.Text("查询状态失败！"),
			),
		)
		return
	}
	if len(stateDatas) < 1 {
		ctx.Send(
			message.ReplyWithMessage(ctx.Event.MessageID,
				message.Text("查询无结果"),
			),
		)
		err = errNoResult
		return
	}
	msg = make(message.Message, 0, 3+len(stateDatas))
	msg = append(msg, message.Reply(ctx.Event.MessageID), message.Text("找到以下状态:\n"))
	for i, state := range stateDatas {
		index := strconv.Itoa(i)
		msg = append(msg, message.Text("["+index+"] "+state.NAME+"\n"))
	}
	return
}
