// Package sgs 三国杀
package sgs

import (
	"strconv"
	"strings"
	"sync"
	"time"

	fcext "github.com/FloatTech/floatbox/ctxext"
	sql "github.com/FloatTech/sqlite"
	ctrl "github.com/FloatTech/zbpctrl"
	"github.com/FloatTech/zbputils/control"
	"github.com/FloatTech/zbputils/ctxext"
	"github.com/sirupsen/logrus"

	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"
)

type sgsdb struct {
	db *sql.Sqlite
	sync.RWMutex
}

var (
	sgsdata = &sgsdb{
		db: &sql.Sqlite{},
	}
	engine = control.Register("sgs", &ctrl.Options[*zero.Ctx]{
		DisableOnDefault: false,
		Brief:            "简易三国杀",
		Help: "蒸蒸日上\n----------指令----------\n" +
			"- 创建房间\n" +
			"- 加入房间\n" +
			"- 退出房间\n" +
			"- 开始游戏\n" +
			"- 查看我的信息\n" +
			"- 查看我的手牌\n" +
			"- 关闭本群所有房间",
		PrivateDataFolder: "sgs",
	})
	getdb = fcext.DoOnceOnSuccess(func(ctx *zero.Ctx) bool {
		sgsdata.db.DBPath = engine.DataFolder() + "sgs.db"
		err := sgsdata.db.Open(time.Hour * 24)
		if err != nil {
			ctx.SendChain(message.Text("[ERROR]:", err))
			return false
		}
		return true
	})
)

func init() {
	engine.OnRegex(`^创建房间$`, getdb).SetBlock(true).Limit(ctxext.LimitByGroup).Handle(func(ctx *zero.Ctx) {
		sgsdata.createRoom(ctx)
	})
	engine.OnRegex(`^加入房间$`, getdb).SetBlock(true).Limit(ctxext.LimitByGroup).Handle(func(ctx *zero.Ctx) {
		sgsdata.joinRoom(ctx)
	})
	engine.OnRegex(`^退出房间$`, getdb).SetBlock(true).Limit(ctxext.LimitByGroup).Handle(func(ctx *zero.Ctx) {
		sgsdata.leftRoom(ctx)
	})
	engine.OnRegex(`^开始游戏$`, getdb).SetBlock(true).Limit(ctxext.LimitByGroup).Handle(func(ctx *zero.Ctx) {
		sgsdata.sgsPlay(ctx)
	})
	engine.OnRegex(`^查看我的信息$`, getdb).SetBlock(true).Limit(ctxext.LimitByUser).Handle(func(ctx *zero.Ctx) {
		sgsdata.findGamerInfo(ctx)
	})
	engine.OnRegex(`^查看我的手牌$`, getdb).SetBlock(true).Limit(ctxext.LimitByUser).Handle(func(ctx *zero.Ctx) {
		sgsdata.findGamerHandCards(ctx)
	})
	engine.OnRegex(`^邀请游戏\s*\[CQ:at,qq=(\d+).*`, getdb).SetBlock(true).Limit(ctxext.LimitByUser).
		Handle(func(ctx *zero.Ctx) {
			sgsdata.inviteGamer(ctx)
		})
	engine.OnRegex("^关闭本群所有房间$", zero.SuperUserPermission, getdb).SetBlock(true).Handle(func(ctx *zero.Ctx) {
		sgsdata.delAllRoom(ctx)
	})
}

// 创建房间
func (sql *sgsdb) createRoom(ctx *zero.Ctx) (err error) {
	sql.Lock()
	defer sql.Unlock()
	err = sql.db.Create("roomInfo", &roomInfo{})
	if err != nil {
		logrus.Errorln("[ERROR] createRoom err: ", err)
		ctx.Send(
			message.ReplyWithMessage(ctx.Event.MessageID,
				message.Text("创建房间失败！"),
			),
		)
		return
	}
	err = sql.db.Create("gameInfo", &gamerInfo{})
	if err != nil {
		logrus.Errorln("[ERROR] createRoom err: ", err)
		ctx.Send(
			message.ReplyWithMessage(ctx.Event.MessageID,
				message.Text("创建角色信息失败！"),
			),
		)
		return
	}
	// 查询是否在房间中
	if sql.db.CanFind("gameInfo", "where GamerID = "+strconv.FormatInt(ctx.Event.UserID, 10)+" and RoomID <> ''") {
		ctx.Send(
			message.ReplyWithMessage(ctx.Event.MessageID,
				message.Text("您已在房间中，无法创建房间！"),
			),
		)
		return
	}
	roomId := time.Now().Unix()
	room := roomInfo{
		RoomID:  roomId,
		HostID:  ctx.Event.UserID,
		GroupID: ctx.Event.GroupID,
		Seats:   strconv.FormatInt(ctx.Event.UserID, 10) + "/" + EMPTY,
		Flag:    "N",
	}
	gamer := gamerInfo{
		GamerID: ctx.Event.UserID,
		GroupID: ctx.Event.GroupID,
		RoomID:  roomId,
		SeatID:  1,
	}
	// TODO 极端情况下两个用户创建同一房间
	sql.db.Insert("gameInfo", &gamer)
	sql.db.Insert("roomInfo", &room)
	ctx.Send(
		message.ReplyWithMessage(ctx.Event.MessageID,
			message.Text("创建房间成功！"),
		),
	)
	return
}

// 加入房间
func (sql *sgsdb) joinRoom(ctx *zero.Ctx) (err error) {
	sql.Lock()
	defer sql.Unlock()
	// 查询是否在房间中
	if sql.db.CanFind("gameInfo", "where GamerID = "+strconv.FormatInt(ctx.Event.UserID, 10)+" and RoomID <> ''") {
		ctx.Send(
			message.ReplyWithMessage(ctx.Event.MessageID,
				message.Text("您已在房间中，无法加入房间！"),
			),
		)
		return
	}
	var roomDatas []roomInfo
	roomData := roomInfo{}
	// 获取房间信息
	err = sql.db.FindFor("roomInfo", &roomData, "where GroupID = "+strconv.FormatInt(ctx.Event.GroupID, 10)+" order by RoomID", func() error {
		if roomData.HostID != 0 {
			roomDatas = append(roomDatas, roomData)
		}
		return nil
	})
	if len(roomDatas) < 1 {
		ctx.Send(
			message.ReplyWithMessage(ctx.Event.MessageID,
				message.Text("当前群没有房间"),
			),
		)
		return
	}
	msg := make(message.Message, 0, 3+len(roomDatas))
	msg = append(msg, message.Reply(ctx.Event.MessageID), message.Text("找到以下房间:\n"))
	for i, info := range roomDatas {
		index := strconv.Itoa(i)
		seatStrIds := strings.Split(info.Seats, "/")
		roomInfoList := make([]string, 0, len(seatStrIds))
		for _, id := range seatStrIds {
			roomInfoList = append(roomInfoList, getGamerNameByIdStr(id, ctx))
		}
		msg = append(msg, message.Text("["+index+"] "+strings.Join(roomInfoList, "/")+"\n"))
	}
	msg = append(msg, message.Text("————————\n输入对应序号加入房间,或回复“取消”取消"))
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
					message.Text("加入房间超时"),
				),
			)
			return
		case m := <-recv:
			cmd := m.Event.Message.String()
			if cmd == "取消" {
				ctx.Send(
					message.ReplyWithMessage(ctx.Event.MessageID,
						message.Text("已取消加入房间"),
					),
				)
				return
			}
			num, _ := strconv.Atoi(cmd)
			if num > len(roomDatas)-1 {
				ctx.Send(
					message.ReplyWithMessage(ctx.Event.MessageID,
						message.Text("房间序号不合法"),
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
	roomData = roomDatas[index]
	// 判断房间是否存在
	if !sql.db.CanFind("roomInfo", "where RoomID = "+strconv.FormatInt(roomData.RoomID, 10)) {
		ctx.Send(
			message.ReplyWithMessage(ctx.Event.MessageID,
				message.Text("房间不存在！"),
			),
		)
		return
	}
	if !in(EMPTY, strings.Split(roomData.Seats, "/")) {
		ctx.Send(
			message.ReplyWithMessage(ctx.Event.MessageID,
				message.Text("房间已满"),
			),
		)
		return
	}
	roomData.Seats = strings.ReplaceAll(roomData.Seats, EMPTY, strconv.FormatInt(ctx.Event.UserID, 10))
	gamer := gamerInfo{
		GamerID: ctx.Event.UserID,
		GroupID: ctx.Event.GroupID,
		RoomID:  roomData.RoomID,
		SeatID:  2,
	}
	sql.db.Insert("roomInfo", &roomData)
	sql.db.Insert("gameInfo", &gamer)
	ctx.Send(
		message.ReplyWithMessage(ctx.Event.MessageID,
			message.Text("加入房间成功"),
		),
	)
	return
}

// 退出房间
func (sql *sgsdb) leftRoom(ctx *zero.Ctx) (err error) {
	sgsData, err := getGameInfo(ctx.Event.UserID, sql)
	if err != nil && err != errGamerNotEnough {
		logrus.Errorln("[ERROR] leftRoom err: ", err)
		ctx.Send(
			message.ReplyWithMessage(ctx.Event.MessageID,
				message.Text("用户信息有误，退出房间失败"),
			),
		)
		return
	}
	// 游戏是开始状态，返回
	if sgsData.roomData.Flag == "Y" {
		ctx.Send(
			message.ReplyWithMessage(ctx.Event.MessageID,
				message.Text("游戏已开始，无法退出房间"),
			),
		)
		return
	}
	roomData := roomInfo{}
	// 不是房主 仅退出房间
	if !sql.db.CanFind("roomInfo", "where HostID = "+strconv.FormatInt(ctx.Event.UserID, 10)) {
		// 删除用户信息
		sql.db.Del("gameInfo", "where GamerID = "+strconv.FormatInt(ctx.Event.UserID, 10))
		err = sql.db.Find("roomInfo", &roomData, "where Seats LIKE '%"+strconv.FormatInt(ctx.Event.UserID, 10)+"%'")
		roomData.Seats = strings.ReplaceAll(roomData.Seats, strconv.FormatInt(ctx.Event.UserID, 10), EMPTY)
		ctx.Send(
			message.ReplyWithMessage(ctx.Event.MessageID,
				message.Text("退出房间成功"),
			),
		)
		sgsData := sgsInfo{
			roomData: roomData,
		}
		updateGameInfo(sgsData, sql)
		return
	}
	err = sql.db.Find("roomInfo", &roomData, "where HostID ="+strconv.FormatInt(ctx.Event.UserID, 10))
	// 删除房间信息
	sql.db.Del("gameInfo", "where GamerID = "+strconv.FormatInt(ctx.Event.UserID, 10))
	sql.db.Del("roomInfo", "where RoomID = "+strconv.FormatInt(roomData.RoomID, 10))
	var msg message.Message
	for _, gamerName := range strings.Split(roomData.Seats, "/") {
		if gamerName == "" || gamerName == EMPTY {
			continue
		}
		gamerId, _ := strconv.ParseInt(gamerName, 10, 64)
		msg = append(msg, message.At(gamerId))
	}
	msg = append(msg, message.Text("\n房主"+ctx.CardOrNickName(ctx.Event.UserID)+"已解散房间"))
	ctx.Send(msg)
	return
}

// 关闭本群所有房间
func (sql *sgsdb) delAllRoom(ctx *zero.Ctx) (err error) {
	if !sql.db.CanFind("roomInfo", "where GroupID = "+strconv.FormatInt(ctx.Event.GroupID, 10)) {
		ctx.Send(
			message.ReplyWithMessage(ctx.Event.MessageID,
				message.Text("本群无创建房间，无需关闭"),
			),
		)
		return
	}
	// 删除房间信息
	sql.db.Del("gameInfo", "where GroupID = "+strconv.FormatInt(ctx.Event.GroupID, 10))
	sql.db.Del("roomInfo", "where GroupID = "+strconv.FormatInt(ctx.Event.GroupID, 10))
	ctx.Send(
		message.ReplyWithMessage(ctx.Event.MessageID,
			message.Text("本群所有房间已关闭"),
		),
	)
	return
}

// // 根据房间创建机器人
// func (sql *sgsdb) createRobots(roomId int64) (message string, err error) {
// 	sql.Lock()
// 	defer sql.Unlock()
// 	roomData := roomInfo{}
// 	err = sql.db.Find("roomInfo", &roomData, "where RoomID = "+strconv.FormatInt(roomId, 10))
// 	if err != nil {
// 		return "创建机器人失败！", err
// 	}
// 	roomSeats := strings.Split(roomData.Seats, "/")
// 	var newRoomSeats []string
// 	for index, gamer := range roomSeats {
// 		var gamerId int64
// 		if gamer == EMPTY {
// 			// TODO 机器人ID
// 			gamerId = int64(806070 + index)
// 		} else {
// 			gamerId, _ = strconv.ParseInt(gamer, 10, 64)
// 		}
// 		gamerData := gameInfo{
// 			GamerID: gamerId,
// 			GroupID: groupId,
// 			RoomID:  roomId,
// 			SeatID:  int64(index),
// 		}
// 		newRoomSeats = append(newRoomSeats, strconv.FormatInt(gamerId, 10))
// 		sql.db.Insert("gameInfo", &gamerData)
// 	}
// 	// 回写最新座位信息
// 	roomData.Seats = strings.Join(newRoomSeats, "/")
// 	err = sql.db.Insert("roomInfo", &roomData)
// 	return "创建机器人成功！", err
// }

// 开始游戏
func (sql *sgsdb) sgsPlay(ctx *zero.Ctx) (err error) {
	sql.Lock()
	defer sql.Unlock()
	// 获取房间信息
	sgsData, err := getGameInfo(ctx.Event.UserID, sql)
	if err != nil {
		logrus.Errorln("[ERROR] sgsPlay err: ", err)
		ctx.Send(
			message.ReplyWithMessage(ctx.Event.MessageID,
				message.Text("您没有可以开始的游戏"),
			),
		)
		return
	}
	if sgsData.roomData.HostID != ctx.Event.UserID {
		ctx.Send(
			message.ReplyWithMessage(ctx.Event.MessageID,
				message.Text("您不是房主，无法开始游戏！"),
			),
		)
		return
	}
	// 本群已有游戏开始
	if sql.db.CanFind("roomInfo", "where GroupID = "+strconv.FormatInt(ctx.Event.GroupID, 10)+" and Flag = 'Y'") {
		ctx.Send(
			message.ReplyWithMessage(ctx.Event.MessageID,
				message.Text("本群已有游戏开始，请稍后再试"),
			),
		)
		return
	}
	ctx.Send(
		message.ReplyWithMessage(ctx.Event.MessageID,
			message.Text("游戏开始"),
		),
	)
	gameData := gameAction{
		UserID: ctx.Event.UserID,
		Ctx:    ctx,
		Sql:    sql,
	}
	// 初始化牌堆 用户摸牌
	initializeCards(gameData)
	over := time.NewTimer(600 * time.Second)
	check, finish := false, false
	gamerIds := getGamerIds(sgsData)
	for {
		for _, userId := range gamerIds {
			gameData.UserID = userId
			select {
			case <-over.C:
				check = true
			default:
				// 回合开始
				startStage(gameData)
				// 判定阶段
				assessStage(gameData)
				// 摸牌阶段
				getCardStage(gameData)
				// 出牌阶段
				err = actionStage(gameData)
				if err == errGamerDie {
					finish = true
					break
				}
				// 弃牌阶段
				abandoningStage(gameData)
				// 回合结束
				endStage(gameData)
			}
			if check || finish {
				break
			}
		}
		if finish {
			break
		}
		if check {
			ctx.SendChain(
				message.At(ctx.Event.UserID),
				message.Text("\n超出最大游戏时间，游戏结束"),
			)
			break
		}
	}
	winer := findWiner(ctx.Event.UserID, sql)
	if winer != (gamerInfo{}) {
		ctx.SendChain(message.At(winer.GamerID), message.Text("\n恭喜获得胜利！"))
	}
	// 游戏结束
	finishGame(ctx.Event.UserID, sql)
	return
}

// 查看我的信息
func (sql *sgsdb) findGamerInfo(ctx *zero.Ctx) (err error) {
	sgsData, err := getGameInfo(ctx.Event.UserID, sql)
	if err != nil {
		logrus.Errorln("[ERROR] findGamerInfo err: ", err)
		ctx.Send(
			message.ReplyWithMessage(ctx.Event.MessageID,
				message.Text("查询用户信息失败"),
			),
		)
		return
	}
	msg := make(message.Message, 0, 5)
	msg = append(msg, message.Text("角色名称: "+sgsData.gamer.Name))
	msg = append(msg, message.Text("\n座位ID: "+strconv.FormatInt(sgsData.gamer.SeatID, 10)))
	msg = append(msg, message.Text("\n血量: "+sgsData.gamer.Blood+"(当前血量/最大血量)"))
	msg = append(msg, message.Text("\n判定区: "+sgsData.gamer.Assessment))
	msg = append(msg, message.Text("\n装备区: "+sgsData.gamer.Equipment))
	ctx.Send(message.ReplyWithMessage(ctx.Event.MessageID, msg...))
	return
}

// 查看我的手牌
func (sql *sgsdb) findGamerHandCards(ctx *zero.Ctx) (err error) {
	sgsData, err := getGameInfo(ctx.Event.UserID, sql)
	if err != nil {
		logrus.Errorln("[ERROR] findGamerHandCards err: ", err)
		ctx.Send(
			message.ReplyWithMessage(ctx.Event.MessageID,
				message.Text("查询用户手牌失败"),
			),
		)
		return
	}
	handCards := getCardList(sgsData.gamer.HandCards)
	if len(handCards) == 0 {
		ctx.Send(
			message.ReplyWithMessage(ctx.Event.MessageID,
				message.Text("用户当前无手牌"),
			),
		)
		return
	}
	msg := make(message.Message, 0, 1+len(handCards))
	msg = append(msg, message.Text("找到如下手牌:"))
	for i, card := range handCards {
		index := strconv.Itoa(i)
		msg = append(msg, message.Text("\n["+index+"] "+card))
	}
	ctx.Send(message.ReplyWithMessage(ctx.Event.MessageID, msg...))
	return
}

func (sql *sgsdb) inviteGamer(ctx *zero.Ctx) (err error) {
	goalId := ctx.State["regex_matched"].([]string)[1]
	// 不能邀请自己
	if goalId == strconv.FormatInt(ctx.Event.UserID, 10) {
		ctx.SendChain(message.Reply(ctx.Event.MessageID), message.Text("邀请失败，不能邀请自己"))
		return
	}
	// 获取用户信息
	sgsData, err := getGameInfo(ctx.Event.UserID, sql)
	if err != nil && err != errGamerNotEnough {
		logrus.Errorln("[ERROR] inviteGamer err: ", err)
		ctx.Send(
			message.ReplyWithMessage(ctx.Event.MessageID,
				message.Text("用户信息有误，邀请失败"),
			),
		)
		return
	}
	if sgsData.gamer.GamerID != sgsData.roomData.HostID {
		ctx.Send(
			message.ReplyWithMessage(ctx.Event.MessageID,
				message.Text("您不是房主，无法邀请用户"),
			),
		)
		return
	}
	// 查询邀请用户是否在房间中
	if sql.db.CanFind("gameInfo", "where GamerID = "+goalId+" and RoomID <> ''") {
		ctx.Send(
			message.ReplyWithMessage(ctx.Event.MessageID,
				message.Text("对方已在房间中，无法邀请用户"),
			),
		)
		return
	}
	inviteGamerId, _ := strconv.ParseInt(goalId, 10, 64)
	ctx.SendChain(
		message.At(inviteGamerId),
		message.Text(ctx.CardOrNickName(ctx.Event.UserID)+"邀请你加入房间,输入确定加入房间,或回复“取消”取消"),
	)
	recv, cancel := zero.NewFutureEvent("message", 999, false, zero.OnlyGroup, zero.RegexRule(`^(确定|取消)$`), zero.CheckGroup(ctx.Event.GroupID), zero.CheckUser(inviteGamerId)).Repeat()
	defer cancel()
	check := false
	over := time.NewTimer(30 * time.Second)
	for {
		select {
		case <-over.C:
			ctx.Send(
				message.ReplyWithMessage(ctx.Event.MessageID,
					message.Text("对方操作超时"),
				),
			)
			return
		case e := <-recv:
			nextcmd := e.Event.Message.String()
			if nextcmd == "取消" {
				ctx.SendChain(
					message.At(int64(inviteGamerId)),
					message.Text("您已拒绝加入房间"),
				)
				return
			}
			if nextcmd == "确定" {
				check = true
			}
		}
		if check {
			break
		}
	}
	if !in(EMPTY, strings.Split(sgsData.roomData.Seats, "/")) {
		ctx.Send(
			message.ReplyWithMessage(ctx.Event.MessageID,
				message.Text("房间已满，邀请失败"),
			),
		)
		return
	}
	sgsData.roomData.Seats = strings.ReplaceAll(sgsData.roomData.Seats, EMPTY, strconv.FormatInt(ctx.Event.UserID, 10))
	gamer := gamerInfo{
		GamerID: inviteGamerId,
		GroupID: sgsData.gamer.GroupID,
		RoomID:  sgsData.roomData.RoomID,
		SeatID:  2,
	}
	sql.db.Insert("roomInfo", &sgsData.roomData)
	sql.db.Insert("gameInfo", &gamer)
	ctx.SendChain(
		message.At(int64(inviteGamerId)),
		message.Text("您已成功加入房间"),
	)
	return
}
