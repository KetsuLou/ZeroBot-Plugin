// Package dailysign 签到系统
package dailysign

import (
	"errors"
	"sync"
	"time"

	sqlite "database/sql"

	fcext "github.com/FloatTech/floatbox/ctxext"
	sql "github.com/FloatTech/sqlite"
	ctrl "github.com/FloatTech/zbpctrl"
	"github.com/FloatTech/zbputils/control"
	"github.com/FloatTech/zbputils/ctxext"
	"github.com/sirupsen/logrus"

	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"
)

var (
	errNoResult = errors.New("no result")
)

type dailydb struct {
	db *sql.Sqlite
	sync.RWMutex
}

type signdb struct {
	db *sql.Sqlite
	sync.RWMutex
}

type path struct {
	PATH string
}

type config struct {
	ID     int64             // ID
	KEY    sqlite.NullString // KEY
	VALUE  sqlite.NullString // VALUE
	MODULE sqlite.NullString // MODULE
	DATE   sqlite.NullString // DATE
}

type state struct {
	NAME   string            // NAME
	ENABLE sqlite.NullString // ENABLE
	STATE  sqlite.NullString // STATE
	TITLE  sqlite.NullString // TITLE
	INFO   sqlite.NullString // INFO
	MODULE sqlite.NullString // MODULE
	NEXT   sqlite.NullString // NEXT
	DATE   sqlite.NullString // DATE
}

var (
	dailysign = &dailydb{
		db: &sql.Sqlite{},
	}
	signsql = &signdb{
		db: &sql.Sqlite{},
	}
	engine = control.Register("dailysign", &ctrl.Options[*zero.Ctx]{
		DisableOnDefault: false,
		Brief:            "签到系统",
		Help: "简易签到\n----------指令----------\n" +
			"- 查看|修改 模块配置\n" +
			"- 查看模块状态",
		PrivateDataFolder: "dailysign",
	})
	getdb = fcext.DoOnceOnSuccess(func(ctx *zero.Ctx) bool {
		dailysign.db.DBPath = engine.DataFolder() + "daily.db"
		err := dailysign.db.Open(time.Hour * 24)
		if err != nil {
			ctx.SendChain(message.Text("[ERROR]:", err))
			return false
		}
		return true
	})
)

func init() {
	engine.OnRegex(`^(查看|修改)\s*模块配置\s*(.*)`, zero.OnlyToMe, zero.SuperUserPermission, getdb).SetBlock(true).Limit(ctxext.LimitByGroup).Handle(func(ctx *zero.Ctx) {
		if ctx.State["regex_matched"].([]string)[1] == "查看" {
			dailysign.queryConfig(ctx)
		}
		if ctx.State["regex_matched"].([]string)[1] == "修改" {
			dailysign.updateConfig(ctx)
		}
	})
	engine.OnRegex(`^(查看|修改)模块状态\s*(.*)`, zero.OnlyToMe, zero.SuperUserPermission, getdb).SetBlock(true).Limit(ctxext.LimitByGroup).Handle(func(ctx *zero.Ctx) {
		if ctx.State["regex_matched"].([]string)[1] == "查看" {
			dailysign.queryState(ctx)
		}
		if ctx.State["regex_matched"].([]string)[1] == "修改" {
			dailysign.updateState(ctx)
		}
	})
}

// 查看配置
func (sql *dailydb) queryConfig(ctx *zero.Ctx) (err error) {
	flag, _ := dailysign.queryCommon(ctx)
	if !flag {
		return
	}
	module := ctx.State["regex_matched"].([]string)[2]
	signsql.queryConfig(ctx, module)
	signsql.db.Close()
	return
}

// 修改配置
func (sql *dailydb) updateConfig(ctx *zero.Ctx) (err error) {
	flag, _ := dailysign.queryCommon(ctx)
	if !flag {
		return
	}
	module := ctx.State["regex_matched"].([]string)[2]
	signsql.updateConfig(ctx, module)
	signsql.db.Close()
	return
}

// 查看状态
func (sql *dailydb) queryState(ctx *zero.Ctx) (err error) {
	flag, _ := dailysign.queryCommon(ctx)
	if !flag {
		return
	}
	module := ctx.State["regex_matched"].([]string)[2]
	signsql.queryState(ctx, module)
	signsql.db.Close()
	return
}

// 修改状态
func (sql *dailydb) updateState(ctx *zero.Ctx) (err error) {
	flag, _ := dailysign.queryCommon(ctx)
	if !flag {
		return
	}
	module := ctx.State["regex_matched"].([]string)[2]
	signsql.updateState(ctx, module)
	signsql.db.Close()
	return
}

func (sql *dailydb) queryCommon(ctx *zero.Ctx) (bool, error) {
	sql.Lock()
	defer sql.Unlock()
	err := sql.db.Create("path", &path{})
	if err != nil {
		logrus.Errorln("[ERROR] createConfig err: ", err)
		ctx.Send(
			message.ReplyWithMessage(ctx.Event.MessageID,
				message.Text("初始化查询失败！"),
			),
		)
		return false, err
	}
	pathData := path{}
	err = sql.db.Find("path", &pathData, "where path IS NOT NULL")
	if err != nil {
		logrus.Errorln("[ERROR] queryCommon err: ", err)
		ctx.Send(
			message.ReplyWithMessage(ctx.Event.MessageID,
				message.Text("查询路径失败！"),
			),
		)
		return false, err
	}
	signsql.db.DBPath = pathData.PATH
	err = signsql.db.Open(time.Hour * 24)
	if err != nil {
		logrus.Errorln("[ERROR] queryCommon err: ", err)
		ctx.Send(
			message.ReplyWithMessage(ctx.Event.MessageID,
				message.Text("查询配置文件失败！"),
			),
		)
		return false, err
	}
	return true, nil
}
