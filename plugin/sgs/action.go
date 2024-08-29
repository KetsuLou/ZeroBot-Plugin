// Package sgs 三国杀
package sgs

import (
	"strconv"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"
)

// autoSelect表示是否存在唯一卡牌时默认选择
func 杀(gameData gameAction, autoSelect bool) (err error) {
	sgsData, _ := getGameInfo(gameData.UserID, gameData.Sql)
	goalInfo := gamerInfo{}
	for _, info := range sgsData.gamerDatas {
		if info.GamerID == gameData.GoalIDList[0] {
			goalInfo = info
			break
		}
	}
	// 找不到用户
	if goalInfo == (gamerInfo{}) {
		gameData.Ctx.SendChain(
			message.At(sgsData.gamer.GamerID),
			message.Text("\n无该用户，请重新操作"),
		)
		return
	}
	// 查询用户使用杀的情况
	dataInfo := gamerDataInfoToStuct(sgsData.gamer.Info)
	if dataInfo.ShaCount >= 1 {
		gameData.Ctx.SendChain(
			message.At(sgsData.gamer.GamerID),
			message.Text("\n已出过杀，不可出杀"),
		)
		return
	}
	// 操作者是否使用杀
	card, err := askGamerCard(gameData, "杀", autoSelect)
	// 未使用杀，返回
	if card == "" {
		return
	}
	sgsGoalData, _ := getGameInfo(gameData.GoalIDList[0], gameData.Sql)
	gameData.Ctx.SendChain(
		message.Text(sgsData.gamer.Name + "对" + sgsGoalData.gamer.Name + "使用" + card),
	)
	// 询问目标角色是否有闪
	gameGoalData := gameData
	gameGoalData.UserID = gameData.GoalIDList[0]
	card, err = askGamerCard(gameGoalData, "闪", false)
	// 未使用闪，处理血量
	if card == "" {
		err = alterBlood(gameData, 1)
	} else {
		gameData.Ctx.SendChain(
			message.Text(sgsGoalData.gamer.Name + "使用" + card),
		)
	}
	// 更新牌堆与用户信息
	updateCardPile(sgsData.gamer.GamerID, gameData.Sql)
	return
}

// userId 表示使用桃的角色 goalId表示加体力的角色 autoSelect表示是否存在唯一卡牌时默认选择
func 桃(gameData gameAction, autoSelect bool) (isUse bool, err error) {
	sgsGoalData, _ := getGameInfo(gameData.GoalIDList[0], gameData.Sql)
	bloodInfo := updateGamerBloodInfo(sgsGoalData.gamer.Blood, 0, 0)
	if bloodInfo[0] >= bloodInfo[1] {
		gameData.Ctx.SendChain(
			message.Text("您的体力已满，无法使用！"),
		)
		return
	}
	// 询问目标角色是否有桃
	isUse = false
	card, err := askGamerCard(gameData, "桃", autoSelect)
	if card == "" {
		return
	}
	isUse = true
	sgsGoalData, _ = getGameInfo(gameData.GoalIDList[0], gameData.Sql)
	bloodInfo = updateGamerBloodInfo(sgsGoalData.gamer.Blood, 1, 0)
	sgsGoalData.gamer.Blood = strconv.FormatInt(bloodInfo[0], 10) + "/" + strconv.FormatInt(bloodInfo[1], 10)
	gameData.Ctx.SendChain(
		message.Text(sgsGoalData.gamer.Name + "回复1点体力，体力值为" + strconv.FormatInt(bloodInfo[0], 10)),
	)
	sgsGoalData.roomData.DiscardPile = strings.Join(getCardList(sgsGoalData.roomData.TempCardPile, sgsGoalData.roomData.DiscardPile), "/")
	sgsGoalData.roomData.TempCardPile = ""
	updateGameInfo(sgsGoalData, gameData.Sql)
	return
}

func 决斗(gameData gameAction) {

}

func 普通锦囊(gameData gameAction, card string, autoSelect bool) (isUse bool, err error) {
	sgsData, _ := getGameInfo(gameData.UserID, gameData.Sql)
	goalInfoList := []gamerInfo{}
	for _, goalId := range gameData.GoalIDList {
		for _, info := range sgsData.gamerDatas {
			if info.GamerID == goalId {
				goalInfoList = append(goalInfoList, info)
				break
			}
		}
	}
	// 找不到所有用户
	if len(goalInfoList) != len(gameData.GoalIDList) {
		gameData.Ctx.SendChain(
			message.At(sgsData.gamer.GamerID),
			message.Text("\n无该用户，请重新操作"),
		)
		return
	}
	// 决斗
	if card == "决斗" {
		决斗(gameData)
	}
	return
}

// 询问使用卡牌 userId表示询问的用户 autoSelect表示是否存在唯一卡牌时默认选择
func askGamerCard(gameData gameAction, askCardType string, autoSelect bool) (card string, err error) {
	sgsData, _ := getGameInfo(gameData.UserID, gameData.Sql)
	card = ""
	var cards []string
	var handCardsIndex []int
	handCards := getCardList(sgsData.gamer.HandCards)
	for index, card := range handCards {
		if checkCardType(card, askCardType) {
			cards = append(cards, card)
			handCardsIndex = append(handCardsIndex, index)
		}
	}
	// 没有可出的卡牌
	if len(cards) == 0 {
		gameData.Ctx.SendChain(
			message.Text(sgsData.gamer.Name + "没有可出的" + askCardType),
		)
		return
	}
	index := 0
	// 若有多张卡牌或非默认选择询问用户
	if len(cards) != 1 || !autoSelect {
		msg := make(message.Message, 0, 3+len(cards))
		msg = append(msg, message.At(sgsData.gamer.GamerID), message.Text("找到以下"+askCardType+":\n"))
		for i, killCard := range cards {
			index := strconv.Itoa(i)
			msg = append(msg, message.Text("["+index+"] "+killCard+"\n"))
		}
		msg = append(msg, message.Text("————————\n输入对应序号使用,或回复“取消”取消"))
		gameData.Ctx.Send(msg)
		recv, cancel := zero.NewFutureEvent("message", 999, false, zero.OnlyGroup, zero.RegexRule(`^(取消|\d+)$`), zero.CheckGroup(sgsData.gamer.GroupID), zero.CheckUser(sgsData.gamer.GamerID)).Repeat()
		defer cancel()
		check := false
		for {
			select {
			case <-time.After(time.Second * 15):
				gameData.Ctx.SendChain(
					message.At(sgsData.gamer.GamerID),
					message.Text("\n操作超时"),
				)
				return
			case e := <-recv:
				nextcmd := e.Event.Message.String()
				if nextcmd == "取消" {
					gameData.Ctx.SendChain(
						message.At(sgsData.gamer.GamerID),
						message.Text("\n已取消使用"),
					)
					return
				}
				index, err = strconv.Atoi(nextcmd)
				if err != nil || index > len(cards)-1 {
					gameData.Ctx.SendChain(
						message.At(sgsData.gamer.GamerID),
						message.Text("\n请输入正确的序号"),
					)
					continue
				}
				check = true
			}
			if check {
				break
			}
		}
	}
	// 更新用户使用卡牌信息
	updateGamerUseCardInfo(sgsData.gamer.GamerID, askCardType, gameData.Sql)
	// 更新临时牌堆信息
	sgsData.roomData.TempCardPile = strings.Join(getCardList(askCardType, sgsData.roomData.TempCardPile), "/")
	card, handCards[handCardsIndex[index]] = handCards[handCardsIndex[index]], ""
	sgsData.gamer.HandCards = strings.Join(getCardList(handCards...), "/")
	updateGameInfo(sgsData, gameData.Sql)
	return
}

// 扣血方法 userId表示当前回合角色ID goalId表示扣血角色ID
func alterBlood(gameData gameAction, number int64) (err error) {
	sgsGoalData, _ := getGameInfo(gameData.GoalIDList[0], gameData.Sql)
	bloodInfo := updateGamerBloodInfo(sgsGoalData.gamer.Blood, -number, 0)
	sgsGoalData.gamer.Blood = strconv.FormatInt(bloodInfo[0], 10) + "/" + strconv.FormatInt(bloodInfo[1], 10)
	gameData.Ctx.SendChain(
		message.Text(sgsGoalData.gamer.Name + "受到" + strconv.FormatInt(number, 10) + "点伤害，体力值" + strconv.FormatInt(bloodInfo[0], 10)),
	)
	updateGameInfo(sgsGoalData, gameData.Sql)
	if bloodInfo[0] > 0 {
		return
	}
	// 执行濒死方法
	err = dying(gameData)
	return
}

func dying(gameData gameAction) (err error) {
	sgsGoalData, _ := getGameInfo(gameData.GoalIDList[0], gameData.Sql)
	// 濒死 求桃方法
	gameData.Ctx.SendChain(
		message.Text(sgsGoalData.gamer.Name + "进入濒死状态"),
	)
	// 以当前角色开始轮询桃
	var newIds []int64
	newGamerIds := getGamerIds(sgsGoalData)
	gamerIndex := 0
	for index, gamerId := range newGamerIds {
		if gamerId == gameData.UserID {
			gamerIndex = index
			break
		}
	}
	newIds = append(newIds, newGamerIds[gamerIndex:]...)
	newIds = append(newIds, newGamerIds[:gamerIndex]...)
	logrus.Infoln("[INFO] alterBlood newIds: ", newIds)
	isDie := true
	bloodInfo := updateGamerBloodInfo(sgsGoalData.gamer.Blood, 0, 0)
	for _, gamerId := range newIds {
		for {
			// 操作者是否使用桃
			gamerData, _ := getGameInfo(gamerId, gameData.Sql)
			bloodInfo = updateGamerBloodInfo(sgsGoalData.gamer.Blood, 0, 0)
			gameData.Ctx.SendChain(
				message.Text(sgsGoalData.gamer.Name + "向" + gamerData.gamer.Name + "求" + strconv.FormatInt(1-bloodInfo[0], 10) + "个桃"),
			)
			gameDataNow := gameData
			gameDataNow.UserID = gamerId
			isUseCard, err := 桃(gameDataNow, false)
			logrus.Infoln("[INFO] alterBlood gamerId: ", gamerId, ", isUserCard is ", isUseCard)
			// 未知错误询问下一个角色
			if err != nil {
				logrus.Errorln("[ERROR] alterBlood err:", err)
				break
			}
			// 未使用桃询问下一个角色
			if !isUseCard {
				break
			}
			// 获取最新扣血角色信息
			sgsGoalData, _ = getGameInfo(gameData.GoalIDList[0], gameData.Sql)
			bloodInfo = updateGamerBloodInfo(sgsGoalData.gamer.Blood, 0, 0)
			// 非濒死状态 退出循环
			if bloodInfo[0] > 0 {
				isDie = false
				break
			}
		}
		if bloodInfo[0] > 0 {
			break
		}
	}
	if isDie {
		err = errGamerDie
		sgsData, _ := getGameInfo(gameData.UserID, gameData.Sql)
		logrus.Infoln("dying bloodInfo: " + strconv.FormatInt(bloodInfo[0], 10) + "/" + strconv.FormatInt(bloodInfo[1], 10))
		gameData.Ctx.SendChain(
			message.Text(sgsData.gamer.Name + "击杀" + sgsGoalData.gamer.Name + "，" + sgsGoalData.gamer.Name + "阵亡"),
		)
	}
	return
}
