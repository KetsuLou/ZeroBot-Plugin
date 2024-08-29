package sgs

import (
	"errors"

	zero "github.com/wdvxdr1123/ZeroBot"
)

var (
	errGamerDie       = errors.New("gamer die")
	errGamerNotEnough = errors.New("gamer not enough")
	errGamerNotFound  = errors.New("gamer not found")
	errRoomNotFound   = errors.New("room not found")
)

const EMPTY = "空"

type sgsInfo struct {
	gamerDatas []gamerInfo // 玩家信息列表
	roomData   roomInfo    // 房间信息列表
	gamer      gamerInfo   //当前玩家
}

type gamerInfo struct {
	GamerID    int64  // 玩家ID
	GroupID    int64  // 群ID
	RoomID     int64  // 房间ID
	SeatID     int64  // 座位ID
	Name       string // 角色名称
	Blood      string // 当前血量/最大血量
	HandCards  string // 手牌区
	Assessment string // 判定区
	Equipment  string // 装备区 武器/防具/防御马/进攻马
	Info       string // 其他信息

}

type roomInfo struct {
	RoomID       int64  // 房间ID
	HostID       int64  // 房主ID
	GroupID      int64  // 群ID
	Seats        string // 座位顺序 GamerID1/GamerID2
	Flag         string // 是否开始游戏，Y是，N否
	CardPile     string // 摸牌堆
	DiscardPile  string // 弃牌堆
	TempCardPile string // 临时区
}

// 游戏事件结构体
type gameAction struct {
	UserID     int64     // 来源ID
	GoalIDList []int64   // 目标ID列表
	Ctx        *zero.Ctx // ctx
	Sql        *sgsdb    // 数据库参数
}

type gamerDataInfo struct {
	ShaCount int `json:"sha"`
}

const STANDARD_CARD = "♠A 闪电/♠A 决斗/♠2 雌雄双股剑/♠2 八卦阵/♠3 顺手牵羊/♠3 过河拆桥/♠4 顺手牵羊/♠4 过河拆桥/♠5 青龙偃月刀/♠5 绝影/♠6 青釭剑/♠6 乐不思蜀/♠7 杀/♠7 南蛮入侵/♠8 杀/♠8 杀/♠9 杀/♠9 杀/♠0 杀/♠0 杀/♠J 无懈可击/♠J 顺手牵羊/♠Q 丈八蛇矛/♠Q 过河拆桥/♠K 大宛/♠K 南蛮入侵/" +
	"♣A 诸葛连弩/♣A 决斗/♣2 杀/♣2 八卦阵/♣3 杀/♣3 过河拆桥/♣4 杀/♣4 过河拆桥/♣5 杀/♣5 的卢/♣6 杀/♣6 乐不思蜀/♣7 杀/♣7 南蛮入侵/♣8 杀/♣8 杀/♣9 杀/♣9 杀/♣0 杀/♣0 杀/♣J 杀/♣J 杀/♣Q 无懈可击/♣Q 借刀杀人/♣K 无懈可击/♣K 借刀杀人/" +
	"♥A 桃园结义/♥A 万箭齐发/♥2 闪/♥2 闪/♥3 桃/♥3 五谷丰登/♥4 桃/♥4 五谷丰登/♥5 麒麟弓/♥5 赤兔/♥6 桃/♥6 乐不思蜀/♥7 桃/♥7 无中生有/♥8 桃/♥8 无中生有/♥9 桃/♥9 无中生有/♥0 杀/♥0 杀/♥J 杀/♥J 无中生有/♥Q 桃/♥Q 过河拆桥/♥K 闪/♥K 爪黄飞电/" +
	"♦A 诸葛连弩/♦A 决斗/♦2 闪/♦2 闪/♦3 闪/♦3 顺手牵羊/♦4 闪/♦4 顺手牵羊/♦5 闪/♦5 贯石斧/♦6 杀/♦6 闪/♦7 杀/♦7 闪/♦8 杀/♦8 闪/♦9 杀/♦9 闪/♦0 杀/♦0 闪/♦J 闪/♦J 闪/♦Q 桃/♦Q 方天画戟/♦K 杀/♦K 紫骍"
const STANDARD_CARD_EX = "♥Q 闪电/♦Q 无懈可击/♣2 仁王盾/♠2 寒冰剑"

// 30张
const STANDARD_CARD_TYPE_SHA = "♠7 杀/♠8 杀/♠8 杀/♠9 杀/♠9 杀/♠0 杀/♠0 杀/♣2 杀/♣3 杀/♣4 杀/♣5 杀/♣6 杀/♣7 杀/♣8 杀/♣8 杀/♣9 杀/♣9 杀/♣0 杀/♣0 杀/♣J 杀/♣J 杀/♥0 杀/♥0 杀/♥J 杀/♦6 杀/♦7 杀/♦8 杀/♦9 杀/♦0 杀/♦K 杀"

// 15张
const STANDARD_CARD_TYPE_SHAN = "♥2 闪/♥2 闪/♥K 闪/♦2 闪/♦2 闪/♦3 闪/♦4 闪/♦5 闪/♦6 闪/♦7 闪/♦8 闪/♦9 闪/♦0 闪/♦J 闪/♦J 闪"

// 8张
const STANDARD_CARD_TYPE_TAO = "♥3 桃/♥4 桃/♥6 桃/♥7 桃/♥8 桃/♥9 桃/♥Q 桃/♦Q 桃"

const AT_QQ_REGEX_RULE = `\s*\[CQ:at,qq=(\d+)\]`
const SHA_REGEX_RULE = `^杀` + AT_QQ_REGEX_RULE
const TAO_REGEX_RULE = `^桃$`
const COMMON_JINNANG_REGEX_RULE = `^(` +
	`决斗` + AT_QQ_REGEX_RULE + `|` +
	`顺手牵羊` + AT_QQ_REGEX_RULE + `|` +
	`过河拆桥` + AT_QQ_REGEX_RULE + `|` +
	`借刀杀人` + AT_QQ_REGEX_RULE + AT_QQ_REGEX_RULE + `|` +
	`五谷丰登` + `|` +
	`桃园结义` + `|` +
	`万箭齐发` + `|` +
	`南蛮入侵` + `|` +
	`无懈可击` + `|` +
	`无中生有` + `)`
const END_REGEX_RULE = `^结束$`
