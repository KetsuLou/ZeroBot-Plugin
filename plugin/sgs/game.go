package sgs

import (
	"math/rand"
	"strconv"
	"strings"
	"time"

	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"
)

// 回合开始
func startStage(gameData gameAction) {

}

// 判定阶段
func assessStage(gameData gameAction) {
	// TODO
	// 获取房间信息
	// getGameInfo(userId, sql)
}

// 摸牌阶段
func getCardStage(gameData gameAction) {
	// 常规摸牌2张
	getCardCommon(gameData, 2)
}

// 出牌阶段
func actionStage(gameData gameAction) (err error) {
	// 获取房间信息
	sgsData, _ := getGameInfo(gameData.UserID, gameData.Sql)
	gameData.Ctx.SendChain(
		message.At(sgsData.gamer.GamerID),
		message.Text("\n出牌阶段开始"),
	)
	杀类型, 杀匹配 := zero.NewFutureEvent("message", 999, false, zero.OnlyGroup, zero.RegexRule(SHA_REGEX_RULE), zero.CheckGroup(sgsData.gamer.GroupID), zero.CheckUser(sgsData.gamer.GamerID)).Repeat()
	桃类型, 桃匹配 := zero.NewFutureEvent("message", 999, false, zero.OnlyGroup, zero.RegexRule(TAO_REGEX_RULE), zero.CheckGroup(sgsData.gamer.GroupID), zero.CheckUser(sgsData.gamer.GamerID)).Repeat()
	普通锦囊类型, 普通锦囊匹配 := zero.NewFutureEvent("message", 999, false, zero.OnlyGroup, zero.RegexRule(COMMON_JINNANG_REGEX_RULE), zero.CheckGroup(sgsData.gamer.GroupID), zero.CheckUser(sgsData.gamer.GamerID)).Repeat()
	结束类型, 结束匹配 := zero.NewFutureEvent("message", 999, false, zero.OnlyGroup, zero.RegexRule(END_REGEX_RULE), zero.CheckGroup(sgsData.gamer.GroupID), zero.CheckUser(sgsData.gamer.GamerID)).Repeat()
	defer 杀匹配()
	defer 桃匹配()
	defer 普通锦囊匹配()
	defer 结束匹配()
	over := time.NewTimer(60 * time.Second)
	check := false
	for {
		select {
		case <-over.C:
			// 出牌阶段超时
			check = true
		case m := <-杀类型:
			cmd := m.Event.Message.String()
			_, goalIds := getCardAndUserIdByStr(cmd)
			gameData.GoalIDList = goalIds
			err = 杀(gameData, true)
			if err == errGamerDie {
				return
			}
			over.Reset(60 * time.Second)
		case <-桃类型:
			gameData.GoalIDList = []int64{gameData.UserID}
			桃(gameData, true)
			over.Reset(60 * time.Second)
		case m := <-普通锦囊类型:
			cmd := m.Event.Message.String()
			card, goalIds := getCardAndUserIdByStr(cmd)
			gameData.GoalIDList = goalIds
			普通锦囊(gameData, card, true)
			over.Reset(60 * time.Second)
		case <-结束类型:
			check = true
		}
		if check {
			break
		}
	}
	return
}

// 弃牌阶段
func abandoningStage(gameData gameAction) {
	sgsData, _ := getGameInfo(gameData.UserID, gameData.Sql)
	handCards := getCardList(sgsData.gamer.HandCards)
	blood, _ := strconv.Atoi(strings.Split(sgsData.gamer.Blood, "/")[0])
	abandonCardsCount := len(handCards) - blood
	if abandonCardsCount <= 0 {
		return
	}
	regex := strings.Repeat("\\d+ ", abandonCardsCount)
	msg := make(message.Message, 0, 3+len(handCards))
	msg = append(msg, message.At(gameData.UserID), message.Text("\n找到以下手牌:\n"))
	for i, card := range handCards {
		index := strconv.Itoa(i)
		msg = append(msg, message.Text("["+index+"] "+card+"\n"))
	}
	msg = append(msg, message.Text("————————\n您需要弃置"+strconv.Itoa(abandonCardsCount)+"张手牌，请选牌(序号用空格分割)"))
	gameData.Ctx.Send(msg)
	recv, cancel := zero.NewFutureEvent("message", 999, false, zero.OnlyGroup, zero.RegexRule(`^(`+regex[:len(regex)-1]+`)$`), zero.CheckGroup(sgsData.gamer.GroupID), zero.CheckUser(sgsData.gamer.GamerID)).Repeat()
	defer cancel()
	chooseIntList := make([]int, 0, abandonCardsCount)
	check := false
	over := time.NewTimer(60 * time.Second)
	for {
		select {
		case <-over.C:
			check = true
		case e := <-recv:
			nextcmd := e.Event.Message.String()
			chooseList := strings.Split(nextcmd, " ")
			for _, cardIndexStr := range chooseList {
				cardIndex, _ := strconv.Atoi(cardIndexStr)
				if cardIndex > len(handCards) {
					gameData.Ctx.SendChain(
						message.At(sgsData.gamer.GamerID),
						message.Text("输入序号不合法"),
					)
					break
				}
				chooseIntList = append(chooseIntList, int(cardIndex))
			}
			if containsDuplicate(chooseIntList) {
				gameData.Ctx.SendChain(
					message.At(sgsData.gamer.GamerID),
					message.Text("序号存在重复"),
				)
			}
			check = true
		}
		if check {
			break
		}
	}
	if len(chooseIntList) == 0 {
		for i := 0; i < abandonCardsCount; i++ {
			chooseIntList = append(chooseIntList, i)
		}
	}
	bloodInfo := updateGamerBloodInfo(sgsData.gamer.Blood, 0, 0)
	newHandCards := make([]string, 0, bloodInfo[0])
	abandonCards := make([]string, 0, abandonCardsCount)
	for index, card := range handCards {
		if !in(index, chooseIntList) {
			newHandCards = append(newHandCards, card)
		} else {
			abandonCards = append(abandonCards, card)
		}
	}
	sgsData.gamer.HandCards = strings.Join(newHandCards, "/")
	sgsData.roomData.DiscardPile = strings.Join(getCardList(strings.Join(abandonCards, "/"), sgsData.roomData.DiscardPile), "/")
	gameData.Ctx.SendChain(
		message.Text(sgsData.gamer.Name + "弃置" + strings.Join(abandonCards, "/")),
	)
	updateGameInfo(sgsData, gameData.Sql)
}

// 回合结束
func endStage(gameData gameAction) {
	sgsData, _ := getGameInfo(gameData.UserID, gameData.Sql)
	// 重置角色信息
	dataInfo := gamerDataInfo{
		ShaCount: 0,
	}
	sgsData.gamer.Info = gamerDataInfoToString(dataInfo)
	updateGameInfo(sgsData, gameData.Sql)
}

// 初始化牌局
func initializeCards(gameData gameAction) {
	cards := STANDARD_CARD_TYPE_SHA + "/" + STANDARD_CARD_TYPE_SHAN + "/" + STANDARD_CARD_TYPE_TAO
	cardsList := getCardList(cards)
	// 重新洗牌后拿牌
	rand.Shuffle(len(cardsList), func(i, j int) {
		cardsList[i], cardsList[j] = cardsList[j], cardsList[i]
	})
	sgsData, _ := getGameInfo(gameData.UserID, gameData.Sql)
	sgsData.roomData.CardPile = strings.Join(cardsList, "/")
	sgsData.roomData.DiscardPile = ""
	// 初始化角色信息
	dataInfo := gamerDataInfo{
		ShaCount: 0,
	}
	for index := range sgsData.gamerDatas {
		sgsData.gamerDatas[index].HandCards = ""
		sgsData.gamerDatas[index].Blood = "2/2"
		sgsData.gamerDatas[index].Equipment = "///"
		sgsData.gamerDatas[index].Name = gameData.Ctx.CardOrNickName(sgsData.gamerDatas[index].GamerID)
		sgsData.gamerDatas[index].Info = gamerDataInfoToString(dataInfo)
		if sgsData.gamerDatas[index].GamerID == sgsData.gamer.GamerID {
			sgsData.gamer = sgsData.gamerDatas[index]
		}
	}
	sgsData.roomData.Flag = "Y"
	updateGameInfo(sgsData, gameData.Sql)
	gamerIds := getGamerIds(sgsData)
	// 初始化用户手牌
	for _, userId := range gamerIds {
		// 常规摸牌2张
		gameDataNow := gameData
		gameDataNow.UserID = userId
		getCardCommon(gameDataNow, 2)
	}
}

// 摸牌通用方法，返回牌数组
func getCardCommon(gameData gameAction, cardCount int) (cards []string) {
	sgsData, _ := getGameInfo(gameData.UserID, gameData.Sql)
	cardPile := getCardList(sgsData.roomData.CardPile)
	discardPile := getCardList(sgsData.roomData.DiscardPile)
	// 牌数不够，重新洗牌后拿牌
	if len(cardPile) < cardCount {
		rand.Shuffle(len(discardPile), func(i, j int) {
			discardPile[i], discardPile[j] = discardPile[j], discardPile[i]
		})
		cardPile = append(cardPile, discardPile...)
		discardPile = []string{}
	}
	cards, cardPile = cardPile[:cardCount], cardPile[cardCount:]
	sgsData.gamer.HandCards = strings.Join(getCardList(strings.Join(cards, "/"), sgsData.gamer.HandCards), "/")
	sgsData.roomData.CardPile = strings.Join(cardPile, "/")
	sgsData.roomData.DiscardPile = strings.Join(discardPile, "/")
	gameData.Ctx.SendChain(
		message.Text(sgsData.gamer.Name + "从摸牌堆获得" + strconv.Itoa(cardCount) + "张牌"),
	)
	updateGameInfo(sgsData, gameData.Sql)
	return
}
