package sgs

import (
	"fmt"
	"reflect"
	"strconv"
	"testing"

	"github.com/agiledragon/gomonkey"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"
)

// func Test杀(t *testing.T) {
// 	killerData := gamerInfo{
// 		GamerID: 111,
// 		GroupID: 111,
// 		RoomID:  111,
// 	}
// 	goalData := gamerInfo{
// 		GamerID: 112,
// 		GroupID: 111,
// 		RoomID:  111,
// 	}
// 	roomData := roomInfo{
// 		RoomID:  111,
// 		HostID:  111,
// 		GroupID: 111,
// 	}
// 	gamers := make([]gamerInfo, 0, 2)
// 	gamers = append(gamers, killerData)
// 	gamers = append(gamers, goalData)

// 	var ctx *zero.Ctx

// 	sgsData := sgsInfo{
// 		gamerDatas: gamers,
// 		roomData:   roomData,
// 		gamer:      killerData,
// 	}
// 	var sql *sgsdb
// 	gameData := gameAction{
// 		Ctx: ctx,
// 		Sql: sql,
// 	}
// 	initializeCards(gameData)
// }

func TestAlterBlood1(t *testing.T) {
	killerData := gamerInfo{
		GamerID: 111,
		GroupID: 111,
		RoomID:  111,
		Blood:   "1/1",
	}
	goalData := gamerInfo{
		GamerID: 112,
		GroupID: 111,
		RoomID:  111,
		Blood:   "1/1",
	}
	roomData := roomInfo{
		RoomID:  111,
		HostID:  111,
		GroupID: 111,
	}
	gamers := make([]gamerInfo, 0, 2)
	gamers = append(gamers, killerData)
	gamers = append(gamers, goalData)

	var ctx *zero.Ctx

	// sgsData := sgsInfo{
	// 	gamerDatas: gamers,
	// 	roomData:  roomData,
	// 	gamer:     killerData,
	// }
	sgsGoalData := sgsInfo{
		gamerDatas: gamers,
		roomData:   roomData,
		gamer:      goalData,
	}
	var sql *sgsdb
	sgsGoalDataMock := gomonkey.ApplyFunc(getGameInfo, func(goalId int64, sql *sgsdb) (sgsInfo, error) {
		return sgsGoalData, nil
	})
	updateGameInfoMock := gomonkey.ApplyFunc(updateGameInfo, func(sgsInfo, *sgsdb) {
	})
	askGamerCardMockWant := []gomonkey.OutputCell{
		{Values: gomonkey.Params{true, nil, false}},
		{Values: gomonkey.Params{true, nil, false}},
	}
	askGamerCardMock := gomonkey.ApplyFuncSeq(askGamerCard, askGamerCardMockWant)
	ctxMock := gomonkey.ApplyMethod(reflect.TypeOf(ctx), "SendChain", func(*zero.Ctx, ...message.MessageSegment) message.MessageID {
		return message.MessageID{}
	})

	defer sgsGoalDataMock.Reset()
	defer updateGameInfoMock.Reset()
	defer askGamerCardMock.Reset()
	defer ctxMock.Reset()
	gameData := gameAction{
		UserID:     111,
		GoalIDList: []int64{112},
		Ctx:        ctx,
		Sql:        sql,
	}
	h := alterBlood(gameData, 1)
	fmt.Println(h)
}

func TestAlterBlood2(t *testing.T) {
	killerData := gamerInfo{
		GamerID: 111,
		GroupID: 111,
		RoomID:  111,
		Blood:   "1/1",
	}
	goalData := gamerInfo{
		GamerID: 112,
		GroupID: 111,
		RoomID:  111,
		Blood:   "1/1",
	}
	goalLossData := goalData
	bloodInfo := updateGamerBloodInfo(goalLossData.Blood, -1, 0)
	goalLossData.Blood = strconv.FormatInt(bloodInfo[0], 10) + "/" + strconv.FormatInt(bloodInfo[1], 10)
	roomData := roomInfo{
		RoomID:  111,
		HostID:  111,
		GroupID: 111,
	}
	gamers := make([]gamerInfo, 0, 2)
	gamers = append(gamers, goalData)
	gamers = append(gamers, killerData)

	var ctx *zero.Ctx

	sgsData := sgsInfo{
		gamerDatas: gamers,
		roomData:   roomData,
		gamer:      killerData,
	}
	sgsGoalData := sgsInfo{
		gamerDatas: gamers,
		roomData:   roomData,
		gamer:      goalData,
	}
	sgsGoalLossBloodData := sgsInfo{
		gamerDatas: gamers,
		roomData:   roomData,
		gamer:      goalLossData,
	}
	var sql *sgsdb
	sgsGoalDataMockWant := []gomonkey.OutputCell{
		{Values: gomonkey.Params{sgsGoalData, nil}},
		{Values: gomonkey.Params{sgsData, nil}},
		{Values: gomonkey.Params{sgsGoalLossBloodData, nil}},
		{Values: gomonkey.Params{sgsGoalLossBloodData, nil}},
		{Values: gomonkey.Params{sgsGoalLossBloodData, nil}},
	}
	sgsGoalDataMock := gomonkey.ApplyFuncSeq(getGameInfo, sgsGoalDataMockWant)
	updateGameInfoMock := gomonkey.ApplyFunc(updateGameInfo, func(sgsInfo, *sgsdb) {
	})
	桃MockWant := []gomonkey.OutputCell{
		{Values: gomonkey.Params{false, nil}},
		{Values: gomonkey.Params{false, nil}},
	}
	桃Mock := gomonkey.ApplyFuncSeq(桃, 桃MockWant)
	ctxMock := gomonkey.ApplyMethod(reflect.TypeOf(ctx), "SendChain", func(*zero.Ctx, ...message.MessageSegment) message.MessageID {
		return message.MessageID{}
	})

	defer sgsGoalDataMock.Reset()
	defer updateGameInfoMock.Reset()
	defer 桃Mock.Reset()
	defer ctxMock.Reset()
	gameData := gameAction{
		UserID:     111,
		GoalIDList: []int64{112},
		Ctx:        ctx,
		Sql:        sql,
	}
	h := alterBlood(gameData, 1)
	fmt.Println(h)
}

func TestLestRoom(t *testing.T) {
	killerData := gamerInfo{
		GamerID: 111,
		GroupID: 111,
		RoomID:  111,
		Blood:   "1/1",
	}
	goalData := gamerInfo{
		GamerID: 112,
		GroupID: 111,
		RoomID:  111,
		Blood:   "1/1",
	}
	roomData := roomInfo{
		RoomID:  111,
		HostID:  111,
		GroupID: 111,
	}
	gamers := make([]gamerInfo, 0, 2)
	gamers = append(gamers, killerData)
	gamers = append(gamers, goalData)

	var ctx *zero.Ctx

	// sgsData := sgsInfo{
	// 	gamerDatas: gamers,
	// 	roomData:  roomData,
	// 	gamer:     killerData,
	// }
	sgsGoalData := sgsInfo{
		gamerDatas: gamers,
		roomData:   roomData,
		gamer:      goalData,
	}
	sgsGoalDataMock := gomonkey.ApplyFunc(getGameInfo, func(goalId int64, sql *sgsdb) (sgsInfo, error) {
		return sgsGoalData, nil
	})
	updateGameInfoMock := gomonkey.ApplyFunc(updateGameInfo, func(sgsInfo, *sgsdb) {
	})
	askGamerCardMockWant := []gomonkey.OutputCell{
		{Values: gomonkey.Params{false, nil}},
		{Values: gomonkey.Params{false, nil}},
	}
	askGamerCardMock := gomonkey.ApplyFuncSeq(askGamerCard, askGamerCardMockWant)
	ctxMock := gomonkey.ApplyMethod(reflect.TypeOf(ctx), "SendChain", func(*zero.Ctx, ...message.MessageSegment) message.MessageID {
		return message.MessageID{}
	})

	defer sgsGoalDataMock.Reset()
	defer updateGameInfoMock.Reset()
	defer askGamerCardMock.Reset()
	defer ctxMock.Reset()
	h := sgsdata.leftRoom(ctx)
	fmt.Println(h)
}

func TestCheckDistance(t *testing.T) {
	killerData := gamerInfo{
		GamerID:   111,
		GroupID:   111,
		RoomID:    111,
		Blood:     "1/1",
		Equipment: "3///",
	}
	data1 := gamerInfo{
		GamerID:   112,
		GroupID:   111,
		RoomID:    111,
		Blood:     "1/1",
		Equipment: "///",
	}
	data2 := gamerInfo{
		GamerID:   112,
		GroupID:   111,
		RoomID:    111,
		Blood:     "1/1",
		Equipment: "///",
	}
	data3 := gamerInfo{
		GamerID:   112,
		GroupID:   111,
		RoomID:    111,
		Blood:     "1/1",
		Equipment: "///",
	}
	goalData := gamerInfo{
		GamerID:   113,
		GroupID:   111,
		RoomID:    111,
		Blood:     "1/1",
		Equipment: "//+1/",
	}
	roomData := roomInfo{
		RoomID:  111,
		HostID:  111,
		GroupID: 111,
	}
	gamers := make([]gamerInfo, 0, 5)
	gamers = append(gamers, data1)
	gamers = append(gamers, killerData)
	gamers = append(gamers, data2)
	gamers = append(gamers, data3)
	gamers = append(gamers, goalData)

	sgsData := sgsInfo{
		gamerDatas: gamers,
		roomData:   roomData,
		gamer:      killerData,
	}
	sgsDataMock := gomonkey.ApplyFunc(getGameInfo, func(goalId int64, sql *sgsdb) (sgsInfo, error) {
		return sgsData, nil
	})

	defer sgsDataMock.Reset()
	var sql *sgsdb
	attack := checkDistance(111, 113, "攻击", sql)
	notAttack := checkDistance(111, 113, "", sql)
	fmt.Println(attack, notAttack)
}

func TestGetUserIdByStr(t *testing.T) {
	h, ids := getCardAndUserIdByStr("杀  [CQ:at,qq=345] ")
	fmt.Println(h, ids)
	r, ids := getCardAndUserIdByStr("借刀杀人  [CQ:at,qq=0848]  [CQ:at,qq=4080] ")
	fmt.Println(r, ids)
}

func Test普通锦囊(t *testing.T) {
	gameData := gameAction{}
	h, _ := 普通锦囊(gameData, "南蛮", false)
	fmt.Println(h)
}
