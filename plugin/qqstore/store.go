// qq商店
package qqstore

import (
	"bytes"
	"encoding/json"
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

type qqstoredb struct {
	db *sql.Sqlite
	sync.RWMutex
}

type userInfo struct {
	UserID int64   // 用户ID
	Money  float64 // 余额
}

type goodInfo struct {
	CID      int64  // cid
	GoodName string // 商品名称
}

const baseURL = "https://157354.svip.people8.net"
const goodsListURL = baseURL + "/api/goods/list"
const PRICE_UP = 0.05

var (
	basePostData = map[string]any{
		"uid":   "157354",
		"token": "ecc554d5dd91dc49e90801d8589950bb",
	}
)

var (
	qqstoredata = &qqstoredb{
		db: &sql.Sqlite{},
	}
	engine = control.Register("qqstore", &ctrl.Options[*zero.Ctx]{
		DisableOnDefault: false,
		Brief:            "菜鸟小商店",
		Help: "菜鸟小商店\n----------指令----------\n" +
			"- 获取商品\n" +
			"- 获取商品[cid]\n" +
			"- 查看商品明细[gid]\n" +
			"- 购买商品[gid] [下单参数1|下单参数2...]\n" +
			"- 查询余额\n" +
			"- 设置商品[商品名称] [cid]",
		PrivateDataFolder: "qqstore",
	}).ApplySingle(ctxext.DefaultSingle)
	getdb = fcext.DoOnceOnSuccess(func(ctx *zero.Ctx) bool {
		qqstoredata.db.DBPath = engine.DataFolder() + "qqstore.db"
		err := qqstoredata.db.Open(time.Hour * 24)
		if err != nil {
			ctx.SendChain(message.Text("[ERROR]:", err))
			return false
		}
		return true
	})
)

func init() {
	engine.OnRegex(`^获取商品\s*(\d*)$`, getdb).SetBlock(true).Limit(ctxext.LimitByGroup).Handle(func(ctx *zero.Ctx) {
		qqstoredata.getGoodsList(ctx)
	})
	engine.OnRegex(`^查看商品明细\s*(\d+)$`, getdb).SetBlock(true).Limit(ctxext.LimitByGroup).Handle(func(ctx *zero.Ctx) {
		qqstoredata.getGoodDetail(ctx)
	})
	engine.OnRegex(`^购买商品\s*(\d+) (\s*\S*)$`, getdb).SetBlock(true).Limit(ctxext.LimitByGroup).Handle(func(ctx *zero.Ctx) {
		qqstoredata.buyGood(ctx)
	})
	engine.OnRegex(`^查询余额$`, getdb).SetBlock(true).Limit(ctxext.LimitByUser).Handle(func(ctx *zero.Ctx) {
		qqstoredata.getUserInfo(ctx)
	})
	engine.OnRegex(`^设置商品\s*(\s*\S*)\s(\d+)`, zero.SuperUserPermission, getdb).SetBlock(true).Handle(func(ctx *zero.Ctx) {
		qqstoredata.setGoodInfo(ctx)
	})
}

func (sql *qqstoredb) getGoodsList(ctx *zero.Ctx) (err error) {
	cid, _ := strconv.ParseInt(ctx.State["regex_matched"].([]string)[1], 10, 64)
	if cid == 0 {
		goods := qqstoredata.getAllGoodsInfo()
		if len(goods) == 0 {
			ctx.Send(
				message.ReplyWithMessage(ctx.Event.MessageID,
					message.Text("未找到商品信息！"),
				),
			)
			return
		}
		msg := make(message.Message, 0, 2+len(goods))
		for _, info := range goods {
			msg = append(msg, message.Text("["+strconv.FormatInt(info.CID, 10)+"]"+info.GoodName+"\n"))
		}
		msg = append(msg, message.Text("————————\n请输入“获取商品 [uid]”查看商品详细说明"))
		ctx.SendChain(
			ctxext.FakeSenderForwardNode(ctx, message.Text("找到以下商品：")),
			ctxext.FakeSenderForwardNode(ctx, msg...))
		return
	}
	body := basePostData
	body["cid"] = cid
	b, err := json.Marshal(body)
	if err != nil {
		return
	}
	response, err := web.PostData(goodsListURL, "application/json", bytes.NewReader(b))
	if err != nil {
		ctx.SendChain(message.Text("ERROR: ", err))
		return
	}
	// ctx.SendChain(message.Reply(ctx.Event.MessageID), message.Text(binary.BytesToString(response)))
	msg := message.Message{}
	goods := binary.BytesToString(response)
	gjson.Get(goods, "list").ForEach(func(key, value gjson.Result) bool {
		msg = append(msg, message.Text("["+value.Get("gid").String()+"]"+value.Get("name").String()+"\n"))
		return true
	})
	if len(msg) != 0 {
		msg = append(msg, message.Text("————————\n请输入“查看商品明细 [gid]”查看商品详细说明"))
		ctx.SendChain(ctxext.FakeSenderForwardNode(ctx, msg...))
	} else {
		ctx.Send(
			message.ReplyWithMessage(ctx.Event.MessageID,
				message.Text("未找到商品信息！"),
			),
		)
	}
	return
}

func (sql *qqstoredb) getGoodDetail(ctx *zero.Ctx) (err error) {
	gid, _ := strconv.ParseInt(ctx.State["regex_matched"].([]string)[1], 10, 64)
	body := basePostData
	body["gid"] = gid
	b, err := json.Marshal(body)
	if err != nil {
		return
	}
	response, err := web.PostData(goodsListURL, "application/json", bytes.NewReader(b))
	if err != nil {
		ctx.SendChain(message.Text("ERROR: ", err))
		return
	}
	goodInfo := binary.BytesToString(response)
	if gjson.Get(goodInfo, "list").String() == "[]" {
		ctx.Send(
			message.ReplyWithMessage(ctx.Event.MessageID,
				message.Text("未找到商品信息！"),
			),
		)
		return
	}
	msg := message.Message{}
	msg = append(msg, message.Text("商品gid: "+gjson.Get(goodInfo, "list.0.gid").String()+"\n"))
	msg = append(msg, message.Text("商品名称: "+gjson.Get(goodInfo, "list.0.name").String()+"\n"))
	msg = append(msg, message.Text("商品价格: "+strconv.FormatFloat(gjson.Get(goodInfo, "list.0.price").Float()+PRICE_UP, 'f', 2, 64)+"\n"))
	var params []string
	for _, param := range gjson.Get(goodInfo, "list.0.inputs").Array() {
		params = append(params, param.String())
	}
	msg = append(msg, message.Text("商品参数: "+strings.Join(params, "|")+"\n"))
	msg = append(msg, message.Text("商品描述: "+gjson.Get(goodInfo, "list.0.desc").String()+"\n"))
	ctx.SendChain(ctxext.FakeSenderForwardNode(ctx, msg...))
	return
}

func (sql *qqstoredb) buyGood(ctx *zero.Ctx) (err error) {

	return
}

func (sql *qqstoredb) getUserInfo(ctx *zero.Ctx) (err error) {
	sql.Lock()
	defer sql.Unlock()
	err = sql.db.Create("userInfo", &userInfo{})
	if err != nil {
		logrus.Errorln("[ERROR] getUserInfo err: ", err)
		ctx.Send(
			message.ReplyWithMessage(ctx.Event.MessageID,
				message.Text("创建用户信息表失败！"),
			),
		)
		return
	}
	user := userInfo{}
	sql.db.Find("userInfo", &user, "where UserID = "+strconv.FormatInt(ctx.Event.UserID, 10))
	// 用户信息为空
	if user == (userInfo{}) {
		user.UserID = ctx.Event.UserID
		user.Money = 0.00
		sql.db.Insert("userInfo", &user)
	}
	ctx.Send(
		message.ReplyWithMessage(ctx.Event.MessageID,
			message.Text("您的余额为: "+strconv.FormatFloat(user.Money, 'f', 2, 64)),
		),
	)
	return
}

func (sql *qqstoredb) setGoodInfo(ctx *zero.Ctx) (err error) {
	sql.Lock()
	defer sql.Unlock()
	err = sql.db.Create("goodInfo", &goodInfo{})
	if err != nil {
		logrus.Errorln("[ERROR] setGood err: ", err)
		ctx.Send(
			message.ReplyWithMessage(ctx.Event.MessageID,
				message.Text("创建商品信息表失败！"),
			),
		)
		return
	}
	goodName := ctx.State["regex_matched"].([]string)[1]
	cid, _ := strconv.ParseInt(ctx.State["regex_matched"].([]string)[2], 10, 64)
	goodInfo := goodInfo{
		CID:      cid,
		GoodName: goodName,
	}
	sql.db.Insert("goodInfo", &goodInfo)
	ctx.Send(
		message.ReplyWithMessage(ctx.Event.MessageID,
			message.Text("添加商品信息成功！"),
		),
	)
	return
}

// 从数据库查询商品信息
func (sql *qqstoredb) getAllGoodsInfo() (goods []goodInfo) {
	sql.Lock()
	defer sql.Unlock()
	goodInfo := goodInfo{}
	sql.db.FindFor("goodInfo", &goodInfo, "WHERE 1 = 1", func() error {
		goods = append(goods, goodInfo)
		return nil
	})
	return
}
