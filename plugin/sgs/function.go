package sgs

import (
	"encoding/json"
	"math"
	"reflect"
	"strconv"
	"strings"

	"github.com/sirupsen/logrus"
	zero "github.com/wdvxdr1123/ZeroBot"
)

const EPSILON = 0.00001

// 判断数字是否重复
func containsDuplicate(nums []int) bool {
	n := make(map[int]int)
	for i := 0; i < len(nums); i++ {
		if _, ok := n[nums[i]]; ok {
			return true
		}
		n[nums[i]] = i
	}
	return false
}

// 判断数组是否存在某元素
func in(target interface{}, array interface{}) bool {
	targetValue := reflect.ValueOf(array)
	switch reflect.TypeOf(array).Kind() {
	case reflect.Slice, reflect.Array:
		for i := 0; i < targetValue.Len(); i++ {
			if targetValue.Index(i).Interface() == target {
				return true
			}
		}
	case reflect.Map:
		if targetValue.MapIndex(reflect.ValueOf(target)).IsValid() {
			return true
		}
	}

	return false
}

// 获取列表类型卡牌
func getCardList(cards ...string) (newCardList []string) {
	var cardsList []string
	for _, card := range cards {
		cardsList = append(cardsList, strings.Split(card, "/")...)
	}
	for _, card := range cardsList {
		if card != "" {
			newCardList = append(newCardList, card)
		}
	}
	return
}

// 获取游戏角色ID
func getGamerIds(sgsData sgsInfo) (gamerIds []int64) {
	for _, info := range sgsData.gamerDatas {
		gamerIds = append(gamerIds, info.GamerID)
	}
	return
}

// id字符串转用户名
func getGamerNameByIdStr(idStr string, ctx *zero.Ctx) string {
	if idStr == "空" {
		return "空"
	}
	gamerId, _ := strconv.ParseInt(idStr, 10, 64)
	return ctx.CardOrNickName(gamerId)
}

// 根据字符串获取卡牌与ID信息
func getCardAndUserIdByStr(cmd string) (card string, ids []int64) {
	res := cmd
	res = strings.ReplaceAll(res, "[", " ")
	res = strings.ReplaceAll(res, "CQ:at,qq=", " ")
	res = strings.ReplaceAll(res, "]", " ")
	strList := strings.Split(res, " ")
	for index, text := range strList {
		if index == 0 {
			card = text
			continue
		}
		if text == "" {
			continue
		}
		goalId, _ := strconv.ParseInt(text, 10, 64)
		ids = append(ids, goalId)
	}
	return
}

// 修改用户血量信息并返回
func updateGamerBloodInfo(bloodStr string, bloodAdd int64, maxBloodAdd int64) (newBlood []int64) {
	bloodList := strings.Split(bloodStr, "/")
	blood, _ := strconv.ParseInt(bloodList[0], 10, 64)
	maxBlood, _ := strconv.ParseInt(bloodList[1], 10, 64)
	blood += bloodAdd
	maxBlood += maxBloodAdd
	newBlood = []int64{blood, maxBlood}
	return
}

// 获取牌局信息
func getGameInfo(userId int64, sql *sgsdb) (sgsData sgsInfo, err error) {
	// 获取用户信息
	gamerData := gamerInfo{}
	err = sql.db.Find("gameInfo", &gamerData, "where GamerID = "+strconv.FormatInt(userId, 10)+" and RoomID <> ''")
	if err != nil {
		err = errGamerNotFound
		return
	}
	sgsData.gamer = gamerData
	// 获取房间信息
	roomData := roomInfo{}
	sql.db.Find("roomInfo", &roomData, "where RoomID = "+strconv.FormatInt(gamerData.RoomID, 10))
	sgsData.roomData = roomData
	// 房间为空
	if roomData == (roomInfo{}) {
		err = errRoomNotFound
		return
	}
	// 玩家未坐满
	if in(EMPTY, strings.Split(roomData.Seats, "/")) {
		err = errGamerNotEnough
		return
	}
	// 获取用户信息
	var gamerDatas []gamerInfo
	err = sql.db.FindFor("gameInfo", &gamerData, "WHERE RoomID = "+strconv.FormatInt(roomData.RoomID, 10)+" order by SeatID", func() error {
		if gamerData.GamerID != 0 {
			gamerDatas = append(gamerDatas, gamerData)
		}
		return nil
	})
	sgsData.gamerDatas = gamerDatas
	return
}

// 更新牌局信息
func updateGameInfo(sgsData sgsInfo, sql *sgsdb) {
	for _, gamer := range sgsData.gamerDatas {
		if gamer != (gamerInfo{}) && gamer.GamerID != sgsData.gamer.GamerID {
			sql.db.Insert("gameInfo", &gamer)
		}
	}
	if sgsData.roomData != (roomInfo{}) {
		sql.db.Insert("roomInfo", &sgsData.roomData)
	}
	if sgsData.gamer != (gamerInfo{}) {
		sql.db.Insert("gameInfo", &sgsData.gamer)
	}
}

// 寻找胜利者
func findWiner(userId int64, sql *sgsdb) (gamer gamerInfo) {
	var gamers []gamerInfo
	sgsData, _ := getGameInfo(userId, sql)
	blood := 0
	for _, info := range sgsData.gamerDatas {
		blood, _ = strconv.Atoi(strings.Split(info.Blood, "/")[0])
		if blood > 0 {
			gamers = append(gamers, info)
		}
	}
	logrus.Infoln("[INFO] findWiner gamers: ", gamers)
	if len(gamers) != 1 {
		logrus.Warnln("[WARN] findWiner gamers is not single")
		return
	}
	gamer = gamers[0]
	return
}

// 结束游戏，删除房间信息
func finishGame(userId int64, sql *sgsdb) {
	sgsData, _ := getGameInfo(userId, sql)
	// 游戏不是开始状态，返回
	if sgsData.roomData.Flag != "Y" {
		logrus.Infoln("[INFO] finishGame: game not start")
		return
	}
	// 删除用户信息
	sql.db.Del("gameInfo", "where RoomID = "+strconv.FormatInt(sgsData.roomData.RoomID, 10))
	// 删除房间信息
	sql.db.Del("roomInfo", "where RoomID = "+strconv.FormatInt(sgsData.roomData.RoomID, 10))
}

// 更新牌堆信息
func updateCardPile(userId int64, sql *sgsdb) {
	sgsData, _ := getGameInfo(userId, sql)
	dataInfo := gamerDataInfoToStuct(sgsData.gamer.Info)
	// 增加杀使用
	dataInfo.ShaCount += 1
	sgsData.gamer.Info = gamerDataInfoToString(dataInfo)
	sgsData.roomData.DiscardPile = strings.Join(getCardList(sgsData.roomData.TempCardPile, sgsData.roomData.DiscardPile), "/")
	sgsData.roomData.TempCardPile = ""
	updateGameInfo(sgsData, sql)
}

// 更新用户使用卡牌信息
func updateGamerUseCardInfo(userId int64, card string, sql *sgsdb) {
	sgsData, _ := getGameInfo(userId, sql)
	dataInfo := gamerDataInfoToStuct(sgsData.gamer.Info)
	// 增加杀使用
	if card == "杀" {
		dataInfo.ShaCount += 1
	}
	sgsData.gamer.Info = gamerDataInfoToString(dataInfo)
	updateGameInfo(sgsData, sql)
}

// 用户其他信息字符串转结构体
func gamerDataInfoToStuct(info string) (dataInfo gamerDataInfo) {
	err := json.Unmarshal([]byte(info), &dataInfo)
	if err != nil {
		logrus.Errorln("[ERROR] gamerDataInfoToStuct err: ", err)
		dataInfo = gamerDataInfo{}
	}
	return
}

// 用户其他信息结构体转字符串
func gamerDataInfoToString(dataInfo gamerDataInfo) string {
	jsonData, _ := json.Marshal(dataInfo)
	return string(jsonData)
}

func filter[T any](slice []T, f func(T) bool) []T {
	var n []T
	for _, e := range slice {
		if f(e) {
			n = append(n, e)
		}
	}
	return n
}

// 距离校验 userId
func checkDistance(userId int64, goalId int64, checkType string, sql *sgsdb) bool {
	sgsData, _ := getGameInfo(userId, sql)
	gamerDatas := filter(sgsData.gamerDatas, func(gamer gamerInfo) bool {
		gamerBlood, _ := strconv.Atoi(strings.Split(gamer.Blood, "/")[0])
		return gamerBlood > 0
	})
	userIndex, goalIndex := 0, 0
	user, goal := gamerInfo{}, gamerInfo{}
	for index, info := range gamerDatas {
		if userId == info.GamerID {
			userIndex = index
			user = info
		}
		if goalId == info.GamerID {
			goalIndex = index
			goal = info
		}
	}
	userEquipmentDistance := getEquipmentDistance(strings.Split(user.Equipment, "/")[3])
	goalEquipmentDistance := getEquipmentDistance(strings.Split(goal.Equipment, "/")[2])
	indexDistance := math.Abs(float64(userIndex - goalIndex))
	if indexDistance > float64(len(gamerDatas))/2.0 {
		indexDistance = float64(len(gamerDatas)) - indexDistance
	}
	userDistance := indexDistance + userEquipmentDistance + goalEquipmentDistance
	// 距离已满足条件 无需后续校验
	if userDistance < 1+EPSILON {
		return true
	}
	// 非攻击操作则 距离校验不通过
	if checkType != "攻击" {
		return false
	}
	userAttackDistance := getEquipmentDistance(strings.Split(user.Equipment, "/")[0])
	if userAttackDistance < 1+EPSILON {
		// 初始化攻击范围
		userAttackDistance = 1.0
	}
	// 判断攻击范围是否大于距离
	return userDistance < userAttackDistance+EPSILON
}

// 获取卡牌距离或攻击范围
func getEquipmentDistance(card string) float64 {
	if strings.Contains(card, "+1") {
		return 1
	}
	if strings.Contains(card, "-1") {
		return -1
	}
	if strings.Contains(card, "1") {
		return 1
	}
	if strings.Contains(card, "2") {
		return 2
	}
	if strings.Contains(card, "3") {
		return 3
	}
	if strings.Contains(card, "4") {
		return 4
	}
	return 0
}

// 查找卡牌类型
func checkCardType(card string, askCardType string) bool {
	if askCardType == "杀" {
		return strings.Contains(STANDARD_CARD_TYPE_SHA, card)
	}
	if askCardType == "闪" {
		return strings.Contains(STANDARD_CARD_TYPE_SHAN, card)

	}
	if askCardType == "桃" {
		return strings.Contains(STANDARD_CARD_TYPE_TAO, card)
	}
	return false
}
