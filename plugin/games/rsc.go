// Package gamesystem 基于zbp的猜歌插件
package gamesystem

import (
	"math/rand"
	"strconv"

	"github.com/FloatTech/AnimeAPI/wallet"
	"github.com/FloatTech/floatbox/math"
	"github.com/FloatTech/zbputils/ctxext"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"

	// 载入游戏系统
	"github.com/FloatTech/ZeroBot-Plugin/plugin/games/gamesystem" // 游戏系统
)

var point = map[string]int{
	"石头": 1,
	"剪刀": 2,
	"布":  3,
}

func init() {
	// 注册游戏信息
	engine, gameManager, err := gamesystem.Register("石头剪刀布", &gamesystem.GameInfo{
		Command: "- 猜拳 [石头|剪刀|布] [金额money]",
		Help:    "简单的小游戏，使用指定金额货币，通过石头、剪刀、布一决胜负！\n玩家选择一种类型进行猜拳，若胜，则获得金额money；若负，则扣除金额money；若平，则不扣除",
		Rewards: "奖励范围在0~money之间",
	})
	if err != nil {
		panic(err)
	}
	engine.OnRegex(`^猜拳[\s]*(石头|剪刀|布)[\s]*(\d+)$`, func(ctx *zero.Ctx) bool {
		if gameManager.PlayIn(ctx.Event.GroupID) {
			return true
		}
		ctx.SendChain(message.Text("游戏已下架,无法游玩"))
		return false
	}).SetBlock(true).Limit(ctxext.LimitByUser).
		Handle(func(ctx *zero.Ctx) {
			uid := ctx.Event.UserID
			money := wallet.GetWalletOf(uid)
			spendMoney := 0
			spendMoneyStr := ctx.State["regex_matched"].([]string)[2]
			if spendMoneyStr != "" {
				number, err := strconv.Atoi(spendMoneyStr)
				if err != nil {
					ctx.SendChain(message.Text("请输入正确的金额"))
					return
				}
				spendMoney = number
			}
			if money-spendMoney < 0 {
				ctx.SendChain(message.Text("你钱包当前只有", money, "杀币,无法完成支付"))
				return
			}
			botchoose := 1 + rand.Intn(3)
			botfinger := ""
			switch botchoose {
			case 1:
				botfinger = "石头"
			case 2:
				botfinger = "剪刀"
			case 3:
				botfinger = "布"
			}
			model := ctx.State["regex_matched"].([]string)[1]
			result := point[model] - botchoose

			// 如果是石头和布的比较，比较值正负取反
			if math.Abs(result) == 2 {
				result = -(result)
			}
			switch {
			case result < 0:
				err := wallet.InsertWalletOf(uid, spendMoney)
				if err == nil {
					ctx.SendChain(message.Reply(ctx.Event.MessageID), message.Text("我出的是", botfinger, "，获得奖励", spendMoney, "杀币"))
					return
				}
			case result > 0:
				err := wallet.InsertWalletOf(uid, -spendMoney)
				if err == nil {
					ctx.SendChain(message.Reply(ctx.Event.MessageID), message.Text("我出的是", botfinger, "，很遗憾你输了，扣除", spendMoney, "杀币"))
					return
				}
			default:
				ctx.SendChain(message.Reply(ctx.Event.MessageID), message.Text("我出的是", botfinger, "，游戏平局"))
			}
		})
}
