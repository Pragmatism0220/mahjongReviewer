package tools

import (
	"encoding/json"
	"fmt"
	"log"
	"math"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/Pragmatism0220/majsoul/message"
	"github.com/golang/protobuf/proto"
)

/* 当前适配版本：v0.10.154.w/code.js */

const NAMEPREF = 1       //2 for english, 1 for sane amount of weeb, 0 for japanese
const VERBOSELOG = false //dump mjs records to output - will make the file too large for tenhou.net/5 viewer
const PRETTY = true      //make the written log somewhat human readable
const SHOWFU = false     //always show fu/han for scoring - even for limit hands

// words that can end up in log, some are mandatory kanji in places
const JPNAME = 0
const RONAME = 1
const ENNAME = 2

var RUNES = map[string][]string{
	/*hand limits*/
	"mangan":        {"満貫", "Mangan ", "Mangan "},
	"haneman":       {"跳満", "Haneman ", "Haneman "},
	"baiman":        {"倍満", "Baiman ", "Baiman "},
	"sanbaiman":     {"三倍満", "Sanbaiman ", "Sanbaiman "},
	"yakuman":       {"役満", "Yakuman ", "Yakuman "},
	"kazoeyakuman":  {"数え役満", "Kazoe Yakuman ", "Counted Yakuman "},
	"kiriagemangan": {"切り上げ満貫", "Kiriage Mangan ", "Rounded Mangan "},
	/*round enders*/
	"agari":          {"和了", "Agari", "Agari"},
	"ryuukyoku":      {"流局", "Ryuukyoku", "Exhaustive Draw"},
	"nagashimangan":  {"流し満貫", "Nagashi Mangan", "Mangan at Draw"},
	"suukaikan":      {"四開槓", "Suukaikan", "Four Kan Abortion"},
	"sanchahou":      {"三家和", "Sanchahou", "Three Ron Abortion"},
	"kyuushukyuuhai": {"九種九牌", "Kyuushu Kyuuhai", "Nine Terminal Abortion"},
	"suufonrenda":    {"四風連打", "Suufon Renda", "Four Wind Abortion"},
	"suuchariichi":   {"四家立直", "Suucha Riichi", "Four Riichi Abortion"},
	/*scoring*/
	"fu":     {"符" /*"Fu",*/, "符", "Fu"},
	"han":    {"飜" /*"Han",*/, "飜", "Han"},
	"points": {"点" /*"Points",*/, "点", "Points"},
	"all":    {"∀", "∀", "∀"},
	"pao":    {"包", "pao", "Responsibility"},
	/*rooms*/
	"tonpuu":     {"東喰", " East", " East"},
	"hanchan":    {"南喰", " South", " South"},
	"friendly":   {"友人戦", "Friendly", "Friendly"},
	"tournament": {"大会戦", "Tournament", "Tournament"},
	"sanma":      {"三", "3-Player ", "3-Player "},
	"red":        {"赤", " Red", " Red Fives"},
	"nored":      {"", " Aka Nashi", " No Red Fives"},
}

var CFG_MODE = map[string][]string{
	/*matchmode*/
	"1":  {"銅の間", "銅の間", "Bronze Room"},
	"2":  {"銅の間", "銅の間", "Bronze Room"},
	"3":  {"銅の間", "銅の間", "Bronze Room"},
	"4":  {"銀の間", "銀の間", "Silver Room"},
	"5":  {"銀の間", "銀の間", "Silver Room"},
	"6":  {"銀の間", "銀の間", "Silver Room"},
	"7":  {"金の間", "金の間", "Gold Room"},
	"8":  {"金の間", "金の間", "Gold Room"},
	"9":  {"金の間", "金の間", "Gold Room"},
	"10": {"玉の間", "玉の間", "Jade Room"},
	"11": {"玉の間", "玉の間", "Jade Room"},
	"12": {"玉の間", "玉の間", "Jade Room"},
	"13": {"乱闘の間", "乱闘の間", "Melee Room"},
	"14": {"乱闘の間", "乱闘の間", "Melee Room"},
	"15": {"王座の間", "王座の間", "Throne Room"},
	"16": {"王座の間", "王座の間", "Throne Room"},
	"17": {"銅の間", "銅の間", "Bronze Room"},
	"18": {"銅の間", "銅の間", "Bronze Room"},
	"19": {"銀の間", "銀の間", "Silver Room"},
	"20": {"銀の間", "銀の間", "Silver Room"},
	"21": {"金の間", "金の間", "Gold Room"},
	"22": {"金の間", "金の間", "Gold Room"},
	"23": {"玉の間", "玉の間", "Jade Room"},
	"24": {"玉の間", "玉の間", "Jade Room"},
	"25": {"王座の間", "王座の間", "Throne Room"},
	"26": {"王座の間", "王座の間", "Throne Room"}, // 没有27、28
	"29": {"交流の間", "交流の間", "Casual Match"},
	"30": {"交流の間", "交流の間", "Casual Match"},
	"31": {"交流の間", "交流の間", "Casual Match"},
	"32": {"交流の間", "交流の間", "Casual Match"},
	"33": {"ドラさんモード", "ドラさんモード", "DoraDorara"},
	"34": {"配牌公開", "配牌公開", "Open Hand Match"},
	"35": {"龍の割目", "龍の割目", "Chaotic Wall Break"},
	"36": {"試練の道", "試練の道", "Path of Trial"},
	"37": {"龙争虎斗", "龙争虎斗", "Long Zheng Hu Dou"},
	"38": {"龙争虎斗", "龙争虎斗", "Long Zheng Hu Dou"},
	"39": {"龙争虎斗", "龙争虎斗", "Long Zheng Hu Dou"},
	"40": {"修羅の戦", "修羅の戦", "Battle of Asura"},
	"41": {"赤血の戦", "赤血の戦", "Bloodshed Skirmish"},
	"42": {"特別対局", "特別対局", "Event Room"},
	"43": {"特別対局", "特別対局", "Event Room"},
	"44": {"明鏡の戦", "明鏡の戦", "Battle of Clairvoyance"},
	"45": {"闇夜の戦", "闇夜の戦", "Battle of Darkness"},
	"46": {"幻界の戦", "幻界の戦", "Dreamland Odyssey"},
}

var CFG_LEVEL = map[string][]string{
	/*level definition*/
	"10101": {"初心★1", "初心★1", "Novice I"},
	"10102": {"初心★2", "初心★2", "Novice II"},
	"10103": {"初心★3", "初心★3", "Novice III"},
	"10201": {"雀士★1", "雀士★1", "Adept I"},
	"10202": {"雀士★2", "雀士★2", "Adept II"},
	"10203": {"雀士★3", "雀士★3", "Adept III"},
	"10301": {"雀傑★1", "雀傑★1", "Expert I"},
	"10302": {"雀傑★2", "雀傑★2", "Expert II"},
	"10303": {"雀傑★3", "雀傑★3", "Expert III"},
	"10401": {"雀豪★1", "雀豪★1", "Master I"},
	"10402": {"雀豪★2", "雀豪★2", "Master II"},
	"10403": {"雀豪★3", "雀豪★3", "Master III"},
	"10501": {"雀聖★1", "雀聖★1", "Saint I"},
	"10502": {"雀聖★2", "雀聖★2", "Saint II"},
	"10503": {"雀聖★3", "雀聖★3", "Saint III"},
	"10601": {"魂天", "魂天", "Celestial"},
	"10701": {"魂天Lv1", "魂天Lv1", "Celestial Lv1"},
	"10702": {"魂天Lv2", "魂天Lv2", "Celestial Lv2"},
	"10703": {"魂天Lv3", "魂天Lv3", "Celestial Lv3"},
	"10704": {"魂天Lv4", "魂天Lv4", "Celestial Lv4"},
	"10705": {"魂天Lv5", "魂天Lv5", "Celestial Lv5"},
	"10706": {"魂天Lv6", "魂天Lv6", "Celestial Lv6"},
	"10707": {"魂天Lv7", "魂天Lv7", "Celestial Lv7"},
	"10708": {"魂天Lv8", "魂天Lv8", "Celestial Lv8"},
	"10709": {"魂天Lv9", "魂天Lv9", "Celestial Lv9"},
	"10710": {"魂天Lv10", "魂天Lv10", "Celestial Lv10"},
	"10711": {"魂天Lv11", "魂天Lv11", "Celestial Lv11"},
	"10712": {"魂天Lv12", "魂天Lv12", "Celestial Lv12"},
	"10713": {"魂天Lv13", "魂天Lv13", "Celestial Lv13"},
	"10714": {"魂天Lv14", "魂天Lv14", "Celestial Lv14"},
	"10715": {"魂天Lv15", "魂天Lv15", "Celestial Lv15"},
	"10716": {"魂天Lv16", "魂天Lv16", "Celestial Lv16"},
	"10717": {"魂天Lv17", "魂天Lv17", "Celestial Lv17"},
	"10718": {"魂天Lv18", "魂天Lv18", "Celestial Lv18"},
	"10719": {"魂天Lv19", "魂天Lv19", "Celestial Lv19"},
	"10720": {"魂天Lv20", "魂天Lv20", "Celestial Lv20"},
	"20101": {"初心★1", "初心★1", "Novice I"},
	"20102": {"初心★2", "初心★2", "Novice II"},
	"20103": {"初心★3", "初心★3", "Novice III"},
	"20201": {"雀士★1", "雀士★1", "Adept I"},
	"20202": {"雀士★2", "雀士★2", "Adept II"},
	"20203": {"雀士★3", "雀士★3", "Adept III"},
	"20301": {"雀傑★1", "雀傑★1", "Expert I"},
	"20302": {"雀傑★2", "雀傑★2", "Expert II"},
	"20303": {"雀傑★3", "雀傑★3", "Expert III"},
	"20401": {"雀豪★1", "雀豪★1", "Master I"},
	"20402": {"雀豪★2", "雀豪★2", "Master II"},
	"20403": {"雀豪★3", "雀豪★3", "Master III"},
	"20501": {"雀聖★1", "雀聖★1", "Saint I"},
	"20502": {"雀聖★2", "雀聖★2", "Saint II"},
	"20503": {"雀聖★3", "雀聖★3", "Saint III"},
	"20601": {"魂天", "魂天", "Celestial"},
	"20701": {"魂天Lv1", "魂天Lv1", "Celestial Lv1"},
	"20702": {"魂天Lv2", "魂天Lv2", "Celestial Lv2"},
	"20703": {"魂天Lv3", "魂天Lv3", "Celestial Lv3"},
	"20704": {"魂天Lv4", "魂天Lv4", "Celestial Lv4"},
	"20705": {"魂天Lv5", "魂天Lv5", "Celestial Lv5"},
	"20706": {"魂天Lv6", "魂天Lv6", "Celestial Lv6"},
	"20707": {"魂天Lv7", "魂天Lv7", "Celestial Lv7"},
	"20708": {"魂天Lv8", "魂天Lv8", "Celestial Lv8"},
	"20709": {"魂天Lv9", "魂天Lv9", "Celestial Lv9"},
	"20710": {"魂天Lv10", "魂天Lv10", "Celestial Lv10"},
	"20711": {"魂天Lv11", "魂天Lv11", "Celestial Lv11"},
	"20712": {"魂天Lv12", "魂天Lv12", "Celestial Lv12"},
	"20713": {"魂天Lv13", "魂天Lv13", "Celestial Lv13"},
	"20714": {"魂天Lv14", "魂天Lv14", "Celestial Lv14"},
	"20715": {"魂天Lv15", "魂天Lv15", "Celestial Lv15"},
	"20716": {"魂天Lv16", "魂天Lv16", "Celestial Lv16"},
	"20717": {"魂天Lv17", "魂天Lv17", "Celestial Lv17"},
	"20718": {"魂天Lv18", "魂天Lv18", "Celestial Lv18"},
	"20719": {"魂天Lv19", "魂天Lv19", "Celestial Lv19"},
	"20720": {"魂天Lv20", "魂天Lv20", "Celestial Lv20"},
}

var CFG_SEX = map[string][]string{
	/*sex of character. 1: female, 2: male*/
	"200001": {"1", "1", "1"},
	"200002": {"1", "1", "1"},
	"200003": {"1", "1", "1"},
	"200004": {"1", "1", "1"},
	"200005": {"1", "1", "1"},
	"200006": {"1", "1", "1"},
	"200007": {"1", "1", "1"},
	"200008": {"1", "1", "1"},
	"200009": {"1", "1", "1"},
	"200010": {"1", "1", "1"},
	"200011": {"2", "2", "2"},
	"200012": {"2", "2", "2"},
	"200013": {"2", "2", "2"},
	"200014": {"2", "2", "2"},
	"200015": {"1", "1", "1"},
	"200016": {"1", "1", "1"},
	"200017": {"1", "1", "1"},
	"200018": {"1", "1", "1"},
	"200019": {"1", "1", "1"},
	"200020": {"1", "1", "1"},
	"200021": {"1", "1", "1"},
	"200022": {"2", "2", "2"},
	"200023": {"2", "2", "2"},
	"200024": {"1", "1", "1"},
	"200025": {"2", "2", "2"},
	"200026": {"1", "1", "1"},
	"200027": {"2", "2", "2"},
	"200028": {"1", "1", "1"},
	"200029": {"1", "1", "1"},
	"200030": {"2", "2", "2"},
	"200031": {"2", "2", "2"},
	"200032": {"1", "1", "1"},
	"200033": {"1", "1", "1"},
	"200034": {"1", "1", "1"},
	"200035": {"1", "1", "1"},
	"200036": {"1", "1", "1"},
	"200037": {"1", "1", "1"},
	"200038": {"1", "1", "1"},
	"200039": {"2", "2", "2"},
	"200040": {"1", "1", "1"},
	"200041": {"1", "1", "1"},
	"200042": {"1", "1", "1"},
	"200043": {"1", "1", "1"},
	"200044": {"1", "1", "1"},
	"200045": {"2", "2", "2"},
	"200046": {"1", "1", "1"},
	"200047": {"2", "2", "2"},
	"200048": {"1", "1", "1"},
	"200049": {"2", "2", "2"},
	"200050": {"2", "2", "2"},
	"200051": {"2", "2", "2"},
	"200052": {"1", "1", "1"},
	"200053": {"1", "1", "1"},
	"200054": {"2", "2", "2"},
	"200055": {"1", "1", "1"},
	"200056": {"2", "2", "2"},
	"200057": {"1", "1", "1"},
	"200058": {"1", "1", "1"},
	"200059": {"1", "1", "1"},
	"200060": {"2", "2", "2"},
	"200061": {"1", "1", "1"},
}

var CFG_YAKU = map[string][]string{
	/*yaku name*/
	"1":    {"門前清自摸和", "門前清自摸和", "Fully Concealed Hand"},
	"2":    {"立直", "立直", "Riichi"},
	"3":    {"槍槓", "槍槓", "Robbing a Kan"},
	"4":    {"嶺上開花", "嶺上開花", "After a Kan"},
	"5":    {"海底摸月", "海底摸月", "Under the Sea"},
	"6":    {"河底撈魚", "河底撈魚", "Under the River"},
	"7":    {"役牌 白", "役牌 白", "White Dragon"},
	"8":    {"役牌 發", "役牌 發", "Green Dragon"},
	"9":    {"役牌 中", "役牌 中", "Red Dragon"},
	"10":   {"役牌:自風牌", "役牌:自風牌", "Seat Wind"},
	"11":   {"役牌:場風牌", "役牌:場風牌", "Prevalent Wind"},
	"12":   {"断幺九", "断幺九", "All Simples"},
	"13":   {"一盃口", "一盃口", "Pure Double Sequence"},
	"14":   {"平和", "平和", "Pinfu"},
	"15":   {"混全帯幺九", "混全帯幺九", "Half Outside Hand"},
	"16":   {"一気通貫", "一気通貫", "Pure Straight"},
	"17":   {"三色同順", "三色同順", "Mixed Triple Sequence"},
	"18":   {"ダブル立直", "ダブル立直", "Double Riichi"},
	"19":   {"三色同刻", "三色同刻", "Triple Triplets"},
	"20":   {"三槓子", "三槓子", "Three Quads"},
	"21":   {"対々和", "対々和", "All Triplets"},
	"22":   {"三暗刻", "三暗刻", "Three Concealed Triplets"},
	"23":   {"小三元", "小三元", "Little Three Dragons"},
	"24":   {"混老頭", "混老頭", "All Terminals and Honors"},
	"25":   {"七対子", "七対子", "Seven Pairs"},
	"26":   {"純全帯幺九", "純全帯幺九", "Fully Outside Hand"},
	"27":   {"混一色", "混一色", "Half Flush"},
	"28":   {"二盃口", "二盃口", "Twice Pure Double Sequence"},
	"29":   {"清一色", "清一色", "Full Flush"},
	"30":   {"一発", "一発", "Ippatsu"},
	"31":   {"ドラ", "ドラ", "Dora"},
	"32":   {"赤ドラ", "赤ドラ", "Red Five"},
	"33":   {"裏ドラ", "裏ドラ", "Ura Dora"},
	"34":   {"抜きドラ", "抜きドラ", "Kita"},
	"35":   {"天和", "天和", "Blessing of Heaven"},
	"36":   {"地和", "地和", "Blessing of Earth"},
	"37":   {"大三元", "大三元", "Big Three Dragons"},
	"38":   {"四暗刻", "四暗刻", "Four Concealed Triplets"},
	"39":   {"字一色", "字一色", "All Honors"},
	"40":   {"緑一色", "緑一色", "All Green"},
	"41":   {"清老頭", "清老頭", "All Terminals"},
	"42":   {"国士無双", "国士無双", "Thirteen Orphans"},
	"43":   {"小四喜", "小四喜", "Four Little Winds"},
	"44":   {"四槓子", "四槓子", "Four Quads"},
	"45":   {"九蓮宝燈", "九蓮宝燈", "Nine Gates"},
	"46":   {"八連荘", "八連荘", "Eight-time East Staying"},
	"47":   {"純正九蓮宝燈", "純正九蓮宝燈", "True Nine Gates"},
	"48":   {"四暗刻単騎", "四暗刻単騎", "Single-wait Four Concealed Triplets"},
	"49":   {"国士無双十三面待ち", "国士無双十三面待ち", "Thirteen-wait Thirteen Orphans"},
	"50":   {"大四喜", "大四喜", "Four Big Winds"},
	"51":   {"燕返し", "燕返し", "Tsubame-gaeshi"},
	"52":   {"槓振り", "槓振り", "Kanburi"},
	"53":   {"十二落抬", "十二落抬", "Shiiaruraotai"},
	"54":   {"五門斉", "五門斉", "Uumensai"},
	"55":   {"三連刻", "三連刻", "Three Chained Triplets"},
	"56":   {"一色三順", "一色三順", "Pure Triple Chow"},
	"57":   {"一筒摸月", "一筒摸月", "Iipinmoyue"},
	"58":   {"九筒撈魚", "九筒撈魚", "Chuupinraoyui"},
	"59":   {"人和", "人和", "Hand of Man"},
	"60":   {"大車輪", "大車輪", "Big Wheels"},
	"61":   {"大竹林", "大竹林", "Bamboo Forest"},
	"62":   {"大数隣", "大数隣", "Numerous Neighbours"},
	"63":   {"石の上にも三年", "石の上にも三年", "Ishinouenimosannen"},
	"64":   {"大七星", "大七星", "Big Seven Stars"},
	"1000": {"根", "根", "Root"},
	"1001": {"嶺上開花", "嶺上開花", "After a Kan"},
	"1002": {"嶺上放銃", "嶺上放銃", "Dealing into Win after Kan"},
	"1003": {"無番和", "無番和", "Yakuless Win"},
	"1004": {"槍槓", "槍槓", "Robbing a Kan"},
	"1005": {"対々和", "対々和", "All Triplets"},
	"1006": {"清一色", "清一色", "Full Flush"},
	"1007": {"七対子", "七対子", "Seven Pairs"},
	"1008": {"帯幺九", "帯幺九", "Terminals in All Sets"},
	"1009": {"金勾釣", "金勾釣", "Single Wait after 4 Triplets"},
	"1010": {"清対", "清対", "Pure Triplets"},
	"1011": {"将対", "将対", "2/5/8 Triplets"},
	"1012": {"龍七対", "龍七対", "Seven Pairs with One Duplicate"},
	"1013": {"清七対", "清七対", "Pure Seven Pairs"},
	"1014": {"清金勾釣", "清金勾釣", "Pure Single Wait after 4 Triplets"},
	"1015": {"清龍七対", "清龍七対", "Pure Seven Pairs with One Duplicate"},
	"1016": {"十八羅漢", "十八羅漢", "Four Quads"},
	"1017": {"清十八羅漢", "清十八羅漢", "Pure Four Quads"},
	"1018": {"天和", "天和", "Blessing of Heaven"},
	"1019": {"地和", "地和", "Blessing of Earth"},
	"1020": {"清幺九", "清幺九", "All Terminals"},
	"1021": {"海底摸月", "海底摸月", "Under the Sea"},
}

// senkinin barai yaku - please don't change, yostar..
const DAISANGEN = 37 //daisangen cfg.fan.fan.map_ index
const DAISUUSHI = 50

const TSUMOGIRI = 60 //tenhou tsumogiri symbol

// global variables - don't touch
var ALLOW_KIRIAGE = false //potentially allow this to be true
var TSUMOLOSSOFF = false  //sanma tsumo loss, is set true for sanma when tsumo loss off

func decodeMessage(data []byte) interface{} {
	type messageWithType struct {
		Name string        `json:"name"`
		Data proto.Message `json:"data"`
	}

	if len(data) != 0 {
		name, data_, err := UnwrapData(data)
		if err != nil {
			return nil
		}

		name = name[1:] // 移除开头的 .
		mt := proto.MessageType(name)
		if mt == nil {
			return nil
		}
		messagePtr := reflect.New(mt.Elem())
		if err := proto.Unmarshal(data_, messagePtr.Interface().(proto.Message)); err != nil {
			return nil
		}

		details := messageWithType{
			Name: name[3:], // 移除开头的 lq.
			Data: messagePtr.Interface().(proto.Message),
		}

		switch details.Name {
		case "GameDetailRecords":
			detailRecords := new(message.GameDetailRecords)
			if err := UnwrapMessage(data, detailRecords); err != nil {
				return new(message.GameDetailRecords)
			}
			return detailRecords
		case "RecordNewRound":
			detailRecords := new(message.RecordNewRound)
			if err := UnwrapMessage(data, detailRecords); err != nil {
				return new(message.RecordNewRound)
			}
			return detailRecords
		case "RecordDiscardTile":
			detailRecords := new(message.RecordDiscardTile)
			if err := UnwrapMessage(data, detailRecords); err != nil {
				return new(message.RecordDiscardTile)
			}
			return detailRecords
		case "RecordDealTile":
			detailRecords := new(message.RecordDealTile)
			if err := UnwrapMessage(data, detailRecords); err != nil {
				return new(message.RecordDealTile)
			}
			return detailRecords
		case "RecordChiPengGang":
			detailRecords := new(message.RecordChiPengGang)
			if err := UnwrapMessage(data, detailRecords); err != nil {
				return new(message.RecordChiPengGang)
			}
			return detailRecords
		case "RecordAnGangAddGang":
			detailRecords := new(message.RecordAnGangAddGang)
			if err := UnwrapMessage(data, detailRecords); err != nil {
				return new(message.RecordAnGangAddGang)
			}
			return detailRecords
		case "RecordBaBei":
			detailRecords := new(message.RecordBaBei)
			if err := UnwrapMessage(data, detailRecords); err != nil {
				return new(message.RecordBaBei)
			}
			return detailRecords
		case "RecordLiuJu":
			detailRecords := new(message.RecordLiuJu)
			if err := UnwrapMessage(data, detailRecords); err != nil {
				return new(message.RecordLiuJu)
			}
			return detailRecords
		case "RecordNoTile":
			detailRecords := new(message.RecordNoTile)
			if err := UnwrapMessage(data, detailRecords); err != nil {
				return new(message.RecordNoTile)
			}
			return detailRecords
		case "RecordHule":
			detailRecords := new(message.RecordHule)
			if err := UnwrapMessage(data, detailRecords); err != nil {
				return new(message.RecordHule)
			}
			return detailRecords
		default:
			log.Printf("Didn't know what to do with %s.\n" + details.Name)
			return nil
		}
	}
	return nil
}

func isContains(array []int, val int) bool {
	for _, e := range array {
		if e == val {
			return true
		}
	}
	return false
}

// flatten array
func flatten(arr interface{}, depth int) ([]interface{}, error) {
	return doFlatten([]interface{}{}, arr, depth, -1)
}

func doFlatten(acc []interface{}, arr interface{}, depth int, current int) ([]interface{}, error) {
	var err error
	switch v := arr.(type) {
	case int, int32, string:
		acc = append(acc, v)
	case []int:
		if depth == current {
			acc = append(acc, v)
		} else {
			for i := range v {
				acc, err = doFlatten(acc, v[i], depth, current+1)
				if err != nil {
					return nil, fmt.Errorf("not []int given")
				}
			}
		}
	case []int32:
		if depth == current {
			acc = append(acc, v)
		} else {
			for i := range v {
				acc, err = doFlatten(acc, v[i], depth, current+1)
				if err != nil {
					return nil, fmt.Errorf("not []int32 given")
				}
			}
		}
	case []string:
		if depth == current {
			acc = append(acc, v)
		} else {
			for i := range v {
				acc, err = doFlatten(acc, v[i], depth, current+1)
				if err != nil {
					return nil, fmt.Errorf("not []string given")
				}
			}
		}
	case []interface{}:
		if depth == current {
			acc = append(acc, v)
		} else {
			for i := range v {
				acc, err = doFlatten(acc, v[i], depth, current+1)
				if err != nil {
					return nil, fmt.Errorf("not []interface{} given")
				}
			}
		}
	default:
		return nil, fmt.Errorf("not a legal input given")
	}
	return acc, nil
}

// pad a to length l with f, needed to pad log for >sanma
func pad_right(a []int32, l int, f int32) []int32 {
	for lenA := len(a); lenA < l; lenA++ {
		a = append(a, f)
	}
	return a
}

// take '2m' and return 2 + 10 etc.
func tm2t(str string) int { //tenhou's tile encoding:
	//   11-19    - 1-9 man
	//   21-29    - 1-9 pin
	//   31-39    - 1-9 sou
	//   41-47    - ESWN WGR
	//   51,52,53 - aka 5 man, pin, sou
	num := int(str[0] - '0')
	tcon := map[uint8]int{'m': 1, 'p': 2, 's': 3, 'z': 4}
	if num != 0 {
		return 10*tcon[str[1]] + num
	} else {
		return 50 + tcon[str[1]]
	}
}

// return normal tile from aka, tenhou rep
func deaka(til int) int { //alternativly - use strings
	if 5 == ^^(til / 10) {
		return 10*(til%10) + (^^(til / 10))
	}
	return til
}

// return aka version of tile
func makeaka(til int) int {
	if 5 == (til % 10) { //is a five (or haku)
		return 10*(til%10) + (^^(til / 10))
	}
	return til //can't be/already is aka
}

// round up to nearest hundred iff TSUMOLOSSOFF == true otherwise return 0
func tlround(x int) int {
	if TSUMOLOSSOFF {
		return int(100 * math.Ceil(float64(x)/100))
	} else {
		return 0
	}
}

// parse mjs hule into tenhou agari list
func parsehule(h *message.HuleInfo, kyoku map[string]interface{}) []interface{} { //tenhou log viewer requires 点, 飜) or 役満) to end strings, rest of scoring string is entirely optional
	//who won, points from (self if tsumo), who won or if pao: who's responsible
	var res []interface{}
	if h.Zimo {
		res = []interface{}{h.Seat, h.Seat, h.Seat}
	} else {
		res = []interface{}{h.Seat, kyoku["ldseat"].(uint32), h.Seat}
	}
	delta := []int32{} //we need to compute the delta ourselves to handle double/triple ron
	points := "0"
	var rp int //riichi stick points, -1 means already taken
	if -1 != kyoku["nriichi"] {
		rp = 1000 * (kyoku["nriichi"].(int) + int(kyoku["round"].([]uint32)[2]))
	} else {
		rp = 0
	}
	hb := 100 * int(kyoku["round"].([]uint32)[1]) //base honba payment

	//sekinin barai logic
	pao := false
	liableseat := -1
	var liablefor uint32 = 0

	if h.Yiman { //only worth checking yakuman hands
		for _, e := range h.Fans {
			if DAISUUSHI == e.Id && (-1 != kyoku["paowind"]) { //daisuushi pao
				pao = true
				liableseat = kyoku["paowind"].(int)
				liablefor += e.Val //realistically can only be liable once
			} else if DAISANGEN == e.Id && (-1 != kyoku["paodrag"]) {
				pao = true
				liableseat = kyoku["paodrag"].(int)
				liablefor += e.Val
			}
		}
	}

	if h.Zimo { //ko-oya payment for non-dealer tsumo
		//delta  = [...new Array(kyoku.nplayers)].map(()=> (-hb - h.point_zimo_xian));
		for i := 0; i < kyoku["nplayers"].(int); i++ {
			delta = append(delta, int32(-hb-int(h.PointZimoXian)-tlround((1/2)*int(h.PointZimoXian))))
		}

		if h.Seat == kyoku["dealerseat"] { //oya tsumo
			delta[h.Seat] = int32(rp + (kyoku["nplayers"].(int)-1)*(hb+int(h.PointZimoXian)) + 2*tlround((1/2)*int(h.PointZimoXian)))
			points = strconv.Itoa(int(h.PointZimoXian) + tlround((1/2)*int(h.PointZimoXian)))
		} else { //ko tsumo
			delta[h.Seat] = int32(rp + hb + int(h.PointZimoQin) + (kyoku["nplayers"].(int)-2)*(hb+int(h.PointZimoXian)) + 2*tlround((1/2)*int(h.PointZimoXian)))
			delta[kyoku["dealerseat"].(uint32)] = int32(-hb - int(h.PointZimoQin) - tlround((1/2)*int(h.PointZimoXian)))
			points = strconv.FormatUint(uint64(h.PointZimoXian), 10) + "-" + strconv.FormatUint(uint64(h.PointZimoQin), 10)
		}
	} else { //ron
		for i := 0; i < kyoku["nplayers"].(int); i++ {
			delta = append(delta, 0)
		}
		delta[h.Seat] = int32(rp + (kyoku["nplayers"].(int)-1)*hb + int(h.PointRong))
		delta[kyoku["ldseat"].(uint32)] = int32(-(kyoku["nplayers"].(int)-1)*hb - int(h.PointRong))
		points = strconv.FormatUint(uint64(h.PointRong), 10)
		kyoku["nriichi"] = -1 //mark the sticks as taken, in case of double ron
	}

	//sekinin barai payments
	//    treat pao as the liable player paying back the other players - safe for multiple yakuman
	const OYA = 0
	const KO = 1
	const RON = 2
	// yakuman scoring table: oya, ko, ron  pays; 1: oya wins; 2: ko wins
	YSCORE := [][]int{{0, 16000, 48000}, {16000, 8000, 32000}}

	if pao {
		res[2] = liableseat //this is how tenhou does it - doesn't really seem to matter to akochan or tenhou.net/5
		if h.Zimo {         //liable player needs to payback n yakuman tsumo payments
			if h.Qinjia { //dealer tsumo
				//should treat tsumo loss as ron, luckily all yakuman values round safely for north bisection
				delta[liableseat] -= int32(2*hb + int(liablefor)*2*YSCORE[OYA][KO] + tlround((1/2)*int(liablefor)*YSCORE[OYA][KO])) // 1? only paying back other ko
				for i, _ := range delta {
					if liableseat != i && h.Seat != uint32(i) && kyoku["nplayers"].(int) >= i {
						delta[i] += int32(hb + int(liablefor)*YSCORE[OYA][KO] + tlround((1/2)*int(liablefor)*(YSCORE[OYA][KO])))
					}
				}
				if 3 == kyoku["nplayers"].(int) { //dealer should get north's payment from liable
					if !TSUMOLOSSOFF {
						delta[h.Seat] += int32(int(liablefor) * YSCORE[OYA][KO])
					}
				}
			} else { //non-dealer tsumo
				delta[liableseat] -= int32((kyoku["nplayers"].(int)-2)*hb + int(liablefor)*(YSCORE[KO][OYA]+YSCORE[KO][KO]) + tlround((1/2)*int(liablefor)*YSCORE[KO][KO])) //^^same 1st, but ko
				for i, _ := range delta {
					if liableseat != i && h.Seat != uint32(i) && kyoku["nplayers"].(int) >= i {
						if kyoku["dealerseat"].(uint32) == uint32(i) {
							delta[i] += int32(hb + int(liablefor)*YSCORE[KO][OYA] + tlround((1/2)*int(liablefor)*YSCORE[KO][KO])) //^^same 1st ...
						} else {
							delta[i] += int32(hb + int(liablefor)*YSCORE[KO][KO] + tlround((1/2)*int(liablefor)*YSCORE[KO][KO])) //^^same 1st ...
						}
					}
				}
			}
		} else { //ron
			//liable seat pays the deal-in seat 1/2 yakuman + full honba
			if h.Qinjia {
				delta[liableseat] -= int32((kyoku["nplayers"].(int)-1)*hb + (1/2)*int(liablefor)*YSCORE[OYA][RON])
				delta[kyoku["ldseat"].(uint32)] += int32((kyoku["nplayers"].(int)-1)*hb + (1/2)*int(liablefor)*YSCORE[OYA][RON])
			} else {
				delta[liableseat] -= int32((kyoku["nplayers"].(int)-1)*hb + (1/2)*int(liablefor)*YSCORE[KO][RON])
				delta[kyoku["ldseat"].(uint32)] += int32((kyoku["nplayers"].(int)-1)*hb + (1/2)*int(liablefor)*YSCORE[KO][RON])
			}
		}
	} //if pao
	//append point symbol
	if h.Zimo && h.Qinjia {
		points += RUNES["points"][JPNAME] + RUNES["all"][NAMEPREF]
	} else {
		points += RUNES["points"][JPNAME]
	}

	//score string
	fuhan := strconv.FormatUint(uint64(h.Fu), 10) + RUNES["fu"][NAMEPREF] + strconv.FormatUint(uint64(h.Count), 10) + RUNES["han"][NAMEPREF]
	if h.Yiman { //yakuman
		if SHOWFU {
			res = append(res, fuhan+RUNES["yakuman"][NAMEPREF]+points)
		} else {
			res = append(res, RUNES["yakuman"][NAMEPREF]+points)
		}
	} else if 13 <= h.Count { //kazoe
		if SHOWFU {
			res = append(res, fuhan+RUNES["kazoeyakuman"][NAMEPREF]+points)
		} else {
			res = append(res, RUNES["kazoeyakuman"][NAMEPREF]+points)
		}
	} else if 11 <= h.Count { //sanbaiman
		if SHOWFU {
			res = append(res, fuhan+RUNES["sanbaiman"][NAMEPREF]+points)
		} else {
			res = append(res, RUNES["sanbaiman"][NAMEPREF]+points)
		}
	} else if 8 <= h.Count { //baiman
		if SHOWFU {
			res = append(res, fuhan+RUNES["baiman"][NAMEPREF]+points)
		} else {
			res = append(res, RUNES["baiman"][NAMEPREF]+points)
		}
	} else if 6 <= h.Count { //haneman
		if SHOWFU {
			res = append(res, fuhan+RUNES["haneman"][NAMEPREF]+points)
		} else {
			res = append(res, RUNES["haneman"][NAMEPREF]+points)
		}
	} else if 5 <= h.Count || (4 <= h.Count && 40 <= h.Fu) || (3 <= h.Count && 70 <= h.Fu) { //mangan
		if SHOWFU {
			res = append(res, fuhan+RUNES["mangan"][NAMEPREF]+points)
		} else {
			res = append(res, RUNES["mangan"][NAMEPREF]+points)
		}
	} else if ALLOW_KIRIAGE && ((4 == h.Count && 30 == h.Fu) || (3 == h.Count && 60 == h.Fu)) { //kiriage
		if SHOWFU {
			res = append(res, fuhan+RUNES["kiriagemangan"][NAMEPREF]+points)
		} else {
			res = append(res, RUNES["kiriagemangan"][NAMEPREF]+points)
		}
	} else { //ordinary hand
		res = append(res, fuhan+points)
	}

	for _, e := range h.Fans {
		yaku := ""
		if JPNAME == NAMEPREF {
			yaku += CFG_YAKU[strconv.FormatUint(uint64(e.Id), 10)][JPNAME]
		} else {
			yaku += CFG_YAKU[strconv.FormatUint(uint64(e.Id), 10)][ENNAME]
		}
		yaku += "("
		if h.Yiman {
			yaku += RUNES["yakuman"][JPNAME]
		} else {
			yaku += (strconv.FormatUint(uint64(e.Val), 10) + RUNES["han"][JPNAME])
		}
		yaku += ")"
		res = append(res, yaku)
	}

	return []interface{}{pad_right(delta, 4, 0.), res}
}

// round information, to be reset every RecordNewRound
var kyoku = map[string]interface{}{}

func kyokuInit(leaf *message.RecordNewRound) { //[kyoku, honba, riichi sticks] - NOTE: 4 mult. works for sanma
	kyoku["nplayers"] = len(leaf.Scores)
	kyoku["round"] = []uint32{4*leaf.Chang + leaf.Ju, leaf.Ben, leaf.Liqibang}
	kyoku["initscores"] = leaf.Scores
	kyoku["initscores"] = pad_right(kyoku["initscores"].([]int32), 4, 0)
	if leaf.Dora != "" {
		kyoku["doras"] = []int{tm2t(leaf.Dora)}
	} else {
		t := []int{}
		for _, e := range leaf.Doras {
			t = append(t, tm2t(e))
		}
		kyoku["doras"] = t
	}
	kyoku["draws"] = [][]interface{}{{}, {}, {}, {}}
	kyoku["discards"] = [][]interface{}{{}, {}, {}, {}}
	t1 := [][]int{{}, {}, {}, {}}
	for i, _ := range kyoku["draws"].([][]interface{}) {
		var rge []string
		switch i {
		case 0:
			rge = leaf.Tiles0
		case 1:
			rge = leaf.Tiles1
		case 2:
			rge = leaf.Tiles2
		case 3:
			rge = leaf.Tiles3
		}
		t2 := []int{}
		for _, e := range rge {
			t2 = append(t2, tm2t(e))
		}
		t1[i] = t2
	}
	kyoku["haipais"] = t1

	//treat the last tile in the dealer's hand as a drawn tile
	kyoku["poppedtile"] = kyoku["haipais"].([][]int)[leaf.Ju][len(kyoku["haipais"].([][]int)[leaf.Ju])-1]
	kyoku["haipais"].([][]int)[leaf.Ju] = kyoku["haipais"].([][]int)[leaf.Ju][0 : len(kyoku["haipais"].([][]int)[leaf.Ju])-1]
	kyoku["draws"].([][]interface{})[leaf.Ju] = append(kyoku["draws"].([][]interface{})[leaf.Ju], kyoku["poppedtile"].(int))
	//information we need, but can't expect in every record
	kyoku["dealerseat"] = leaf.Ju
	kyoku["ldseat"] = -1 //who dealt the last tile
	kyoku["nriichi"] = 0 //number of current riichis - needed for scores, abort workaround
	kyoku["nkan"] = 0    //number of current kans - only for abort workaround
	//pao rule
	kyoku["nowinds"] = []int{0, 0, 0, 0} //counter for each players open wind pons/kans
	kyoku["nodrags"] = []int{0, 0, 0, 0}
	kyoku["paowind"] = -1 //seat of who dealt the final wind, -1 if no one is responsible
	kyoku["paodrag"] = -1
}

// dump round informaion
func kyokuDump(uras []int) []interface{} { //NOTE: doras,uras are the indicators
	entry := []interface{}{}
	entry = append(entry, kyoku["round"])
	entry = append(entry, kyoku["initscores"])
	entry = append(entry, kyoku["doras"])
	entry = append(entry, uras)

	for i, f := range kyoku["haipais"].([][]int) {
		entry = append(entry, f)
		entry = append(entry, kyoku["draws"].([][]interface{})[i])
		entry = append(entry, kyoku["discards"].([][]interface{})[i])
	}
	return entry
}

// sekinin barai tiles
var WINDS = []int{41, 42, 43, 44}
var DRAGS = []int{45, 46, 47, 54} //0z would be aka haku

// senkinin barai incrementer - to be called every pon, daiminkan, ankan
func kyokuCountpao(tile, owner, feeder int) { //owner and feeder are seats, tile should be tenhou
	if isContains(WINDS, tile) {
		kyoku["nowinds"].([]int)[owner] = kyoku["nowinds"].([]int)[owner] + 1
		if 4 == kyoku["nowinds"].([]int)[owner] {
			kyoku["paowind"] = feeder
		}
	} else if isContains(DRAGS, tile) {
		kyoku["nodrags"].([]int)[owner] = kyoku["nodrags"].([]int)[owner] + 1
		if 3 == kyoku["nodrags"].([]int)[owner] {
			kyoku["paodrag"] = feeder
		}
	}
	return
}

// seat1 is seat0's x
func relativeseating(seat0, seat1 uint32) uint32 { //0: kamicha, 1: toimen, 2: if shimocha
	return (seat0 - seat1 + 4 - 1) % 4
}

// convert mjs records to tenhou log
func generatelog(mjslog []interface{}) interface{} {
	resLog := []interface{}{} // In the downloadlog.js file, this variable is named "log". However, it conflicts with the package name "log", so it is renamed to "resLog".
	for _, v := range mjslog {
		switch e := v.(type) {
		case *message.RecordNewRound:
			kyokuInit(e)
		case *message.RecordDiscardTile: //discard - marking tsumogiri and riichi
			var symbol interface{}
			if e.Moqie {
				symbol = TSUMOGIRI
			} else {
				symbol = tm2t(e.Tile)
			}
			//we pretend that the dealer's initial 14th tile is drawn - so we need to manually check the first discard
			if e.Seat == kyoku["dealerseat"] && len(kyoku["discards"].([][]interface{})[e.Seat]) == 0 && symbol == kyoku["poppedtile"] {
				symbol = TSUMOGIRI
			}
			if e.IsLiqi { //riichi delcaration
				kyoku["nriichi"] = kyoku["nriichi"].(int) + 1
				symbol = "r" + strconv.Itoa(symbol.(int))
			}
			kyoku["discards"].([][]interface{})[e.Seat] = append(kyoku["discards"].([][]interface{})[e.Seat], symbol)
			kyoku["ldseat"] = e.Seat //for ron, pon etc.

			//sometimes we get dora passed here
			if len(e.Doras) != 0 && (len(e.Doras) > len(kyoku["doras"].([]int))) {
				t := []int{}
				for _, ele := range e.Doras {
					t = append(t, tm2t(ele))
				}
				kyoku["doras"] = t
			}
		case *message.RecordDealTile: //draw - after kan this gets passed the new dora
			if len(e.Doras) != 0 && (len(e.Doras) > len(kyoku["doras"].([]int))) {
				t := []int{}
				for _, ele := range e.Doras {
					t = append(t, tm2t(ele))
				}
				kyoku["doras"] = t
			}
			kyoku["draws"].([][]interface{})[e.Seat] = append(kyoku["draws"].([][]interface{})[e.Seat], tm2t(e.Tile))
		case *message.RecordChiPengGang: //call - chi, pon, daiminkan
			switch e.Type {
			case 0: //chii
				kyoku["draws"].([][]interface{})[e.Seat] = append(kyoku["draws"].([][]interface{})[e.Seat], "c"+strconv.Itoa(tm2t(e.Tiles[2]))+strconv.Itoa(tm2t(e.Tiles[0]))+strconv.Itoa(tm2t(e.Tiles[1])))
			case 1: //pon
				worktiles := []string{}
				for _, ele := range e.Tiles {
					worktiles = append(worktiles, strconv.Itoa(tm2t(ele)))
				}
				idx := relativeseating(e.Seat, kyoku["ldseat"].(uint32))
				tmp, _ := strconv.Atoi(worktiles[0])
				kyokuCountpao(tmp, int(e.Seat), int(kyoku["ldseat"].(uint32)))
				//pop the called tile a preprend 'p'
				// 复制切片而避免修改共享的地址数据后，原来的切片指针指向的数据跟着改变的bug
				left := make([]string, idx)
				for i := 0; i < int(idx); i++ {
					left[i] = worktiles[i]
				}
				right := make([]string, len(worktiles)-int(idx)-1)
				for i := 0; i < len(worktiles)-int(idx)-1; i++ {
					right[i] = worktiles[int(idx)+i]
				}
				last := worktiles[len(worktiles)-1]

				worktiles = append(left, "p"+last) //pop the called tile a preprend 'p'
				worktiles = append(worktiles, right...)

				s := ""
				for _, ele := range worktiles {
					s += ele
				}
				kyoku["draws"].([][]interface{})[e.Seat] = append(kyoku["draws"].([][]interface{})[e.Seat], s)
			case 2:
				// kan naki:
				//   daiminkan:
				//     kamicha   "m39393939" (0)
				//     toimen    "39m393939" (1)
				//     shimocha  "222222m22" (3)
				//     (writes to draws; 0 to discards)
				//   shouminkan: (same order as pon; immediate tile after k is the added tile)
				//     kamicha   "k37373737" (0)
				//     toimen    "31k313131" (1)
				//     shimocha  "3737k3737" (2)
				//     (writes to discards)
				//   ankan:
				//     "121212a12" (3)
				//     (writes to discards)
				///////////////////////////////////////////////////
				//daiminkan
				calltiles := []string{}
				for _, ele := range e.Tiles {
					calltiles = append(calltiles, strconv.Itoa(tm2t(ele)))
				}
				// < kamicha 0 | toimen 1 | shimocha 3 >
				idx := relativeseating(e.Seat, kyoku["ldseat"].(uint32))
				tmp, _ := strconv.Atoi(calltiles[0])
				kyokuCountpao(tmp, int(e.Seat), int(kyoku["ldseat"].(uint32)))
				if idx == 2 {
					left := make([]string, 3)
					for i := 0; i < 3; i++ {
						left[i] = calltiles[i]
					}
					right := make([]string, len(calltiles)-4)
					for i := 0; i < len(calltiles)-4; i++ {
						right[i] = calltiles[3+i]
					}
					last := calltiles[len(calltiles)-1]

					calltiles = append(left, "m"+last)
					calltiles = append(calltiles, right...)
				} else {
					left := make([]string, idx)
					for i := 0; i < int(idx); i++ {
						left[i] = calltiles[i]
					}
					right := make([]string, len(calltiles)-int(idx)-1)
					for i := 0; i < len(calltiles)-int(idx)-1; i++ {
						right[i] = calltiles[int(idx)+i]
					}
					last := calltiles[len(calltiles)-1]

					calltiles = append(left, "m"+last)
					calltiles = append(calltiles, right...)
				}
				s := ""
				for _, ele := range calltiles {
					s += ele
				}
				kyoku["draws"].([][]interface{})[e.Seat] = append(kyoku["draws"].([][]interface{})[e.Seat], s)

				//tenhou drops a 0 in discards for this
				kyoku["discards"].([][]interface{})[e.Seat] = append(kyoku["discards"].([][]interface{})[e.Seat], 0)
				//register kan
				kyoku["nkan"] = kyoku["nkan"].(int) + 1
			default:
				log.Printf("Didn't know what to do with %d.\n", e.Type)
			}
		case *message.RecordAnGangAddGang: //kan - shouminkan 'k', ankan 'a'
			//NOTE: e.tiles here is a single tile; naki is placed in discards
			til := tm2t(e.Tiles)
			kyoku["ldseat"] = e.Seat // for chankan, no conflict as last discard has passed
			switch e.Type {
			case 3: //ankan
				////////////////////
				// mjs chun ankan example record:
				//{"seat":0,"type":3,"tiles":"7z"}
				////////////////////
				kyokuCountpao(til, int(e.Seat), -1) //count the group as visible, but don't set pao
				//get the tiles from haipai and draws that
				//are involved in ankan, dumb
				//because n aka might be involved
				t1 := []int{}
				for _, ele := range kyoku["haipais"].([][]int)[e.Seat] {
					if deaka(ele) == deaka(til) {
						t1 = append(t1, ele)
					}
				}
				t2 := []int{}
				for _, ele := range kyoku["draws"].([][]interface{})[e.Seat] {
					switch val := ele.(type) {
					case int:
						if deaka(val) == deaka(til) {
							t2 = append(t2, val)
						}
					}
				}
				ankantiles := append(t1, t2...)
				til = ankantiles[len(ankantiles)-1] //doesn't really matter which tile we mark ankan with - chosing last drawn
				ankantiles = ankantiles[0 : len(ankantiles)-1]
				s := ""
				for _, ele := range ankantiles {
					s += strconv.Itoa(ele)
				}
				kyoku["discards"].([][]interface{})[e.Seat] = append(kyoku["discards"].([][]interface{})[e.Seat], s+"a"+strconv.Itoa(til)) //push naki
				kyoku["nkan"] = kyoku["nkan"].(int) + 1
			case 2: //shouminkan
				//get pon naki from .draws and swap in new symbol
				t := []string{}
				for _, ele := range kyoku["draws"].([][]interface{})[e.Seat] {
					switch val := ele.(type) {
					case string:
						if strings.Contains(val, "p"+strconv.Itoa(deaka(til))) || strings.Contains(val, "p"+strconv.Itoa(makeaka(til))) { // naki. pon involves same tile type
							t = append(t, val)
						}
					}
				}
				nakis := t
				kyoku["discards"].([][]interface{})[e.Seat] = append(kyoku["discards"].([][]interface{})[e.Seat], strings.Replace(nakis[0], "p", "k"+strconv.Itoa(til), 1)) //push naki
				kyoku["nkan"] = kyoku["nkan"].(int) + 1
			default:
				log.Printf("Didn't know what to do with %d.\n", e.Type)
			}
		case *message.RecordBaBei:
			//kita - this record (only) gives {seat, moqie}
			//NOTE: tenhou doesn't mark its kita based on when they were drawn, so we won't
			//if (e.moqie)
			//    kyoku.discards[e.seat].push("f" + TSUMOGIRI);
			//else
			kyoku["discards"].([][]interface{})[e.Seat] = append(kyoku["discards"].([][]interface{})[e.Seat], "f44")
		/////////////////////////////////////////////////////
		// round enders:
		// "RecordNoTile" - ryuukyoku
		// "RecordHule"   - agari - ron/tsumo
		// "RecordLiuJu"  - abortion
		//////////////////////////////////////////////////////
		case *message.RecordLiuJu: //abortion
			entry := kyokuDump([]int{})
			if 1 == e.Type {
				entry = append(entry, []interface{}{RUNES["kyuushukyuuhai"][NAMEPREF]}) //kyuushukyuhai
			} else if 2 == e.Type {
				entry = append(entry, []interface{}{RUNES["suufonrenda"][NAMEPREF]}) //suufon renda
			} else if 4 == kyoku["nriichi"] { //TODO: actually get the type code
				entry = append(entry, []interface{}{RUNES["suuchariichi"][NAMEPREF]}) //4 riichi
			} else if 4 <= kyoku["nkan"].(int) { //TODO: actually get type code
				entry = append(entry, []interface{}{RUNES["suukaikan"][NAMEPREF]}) //4 kan, potentially false positive on 3 ron with 4 kans
			} else {
				entry = append(entry, []interface{}{RUNES["sanchahou"][NAMEPREF]}) //3 ron - can't actually get this in mjs
			}
			resLog = append(resLog, entry)
		case *message.RecordNoTile: //ryuukyoku
			entry := kyokuDump([]int{})
			delta := []int32{0, 0, 0, 0}

			//NOTE: mjs wll not give delta_scores if everyone is (no)ten - TODO: minimize the autism
			if len(e.Scores) != 0 && e.Scores[0] != nil && len(e.Scores[0].DeltaScores) != 0 {
				for _, f := range e.Scores {
					for i, g := range f.DeltaScores {
						delta[i] += g //for the rare case of multiple nagashi, we sum the arrays
					}
				}
			}

			if e.Liujumanguan { //nagashi mangan
				entry = append(entry, []interface{}{RUNES["nagashimangan"][NAMEPREF], delta})
			} else { //normal ryuukyoku
				entry = append(entry, []interface{}{RUNES["ryuukyoku"][NAMEPREF], delta})
			}
			resLog = append(resLog, entry)
		case *message.RecordHule: //agari
			agari := []interface{}{}
			ura := []int{}
			for _, f := range e.Hules {
				if len(ura) < len(f.LiDoras) { //take the longest ura list - double ron with riichi + dama
					for _, g := range f.LiDoras {
						ura = append(ura, tm2t(g))
					}
				}
				agari = append(agari, parsehule(f, kyoku))
			}

			entry := kyokuDump(ura)

			tstr := []interface{}{RUNES["agari"][JPNAME]}
			flatS, _ := flatten(agari, 1)
			tstr = append(tstr, flatS...)

			entry = append(entry, tstr) //needs the japanese agari
			resLog = append(resLog, entry)
		default:
			log.Printf("Didn't know what to do.\n")
		}
	}
	return resLog
}

func parse(record *message.ResGameRecord) map[string]interface{} {
	res := map[string]interface{}{}
	ruledisp := ""
	lobby := "" //usually 0, is the custom lobby number
	nplayers := uint32(len(record.Head.Result.Players))
	nakas := nplayers - 1 //default
	// anon edit 1 start
	var mjslog []interface{}
	mjsact := decodeMessage(record.Data).(*message.GameDetailRecords).Actions

	for _, e := range mjsact {
		if len(e.Result) != 0 {
			mjslog = append(mjslog, decodeMessage(e.Result))
		}
	}
	// anon edit 1 end

	res["ver"] = "2.3"            // mlog version number
	res["ref"] = record.Head.Uuid // game id - copy and paste into "other" on the log page to view
	res["log"] = generatelog(mjslog)
	//PF4 is yonma, PF3 is sanma
	res["ratingc"] = "PF" + strconv.FormatUint(uint64(nplayers), 10)

	//rule display
	if 3 == nplayers && JPNAME == NAMEPREF {
		ruledisp += RUNES["sanma"][JPNAME]
	}

	// TODO: 这里没有解析code.js，而是偷懒把map自己扒了下来。这种做法当版本更新添加新的键值对（比如新角色、新皮肤等）之后就会失效。
	// TODO: 临时的解决方案是设定一个初始值，比如角色默认设置为一姬，段位默认为初心等。（尚未完成）
	// TODO: 最终的解决方案应该是搞懂code.js的接口运作方式（非常奇怪，搜不到关键词的响应，也抓不到包），然后传值进去，从中请求map等函数。
	if record.Head.Config.Meta.ModeId != 0 { //ranked or casual.银之间、金之间之类的。
		if JPNAME == NAMEPREF {
			ruledisp += CFG_MODE[strconv.FormatUint(uint64(record.Head.Config.Meta.ModeId), 10)][JPNAME]
		} else {
			ruledisp += CFG_MODE[strconv.FormatUint(uint64(record.Head.Config.Meta.ModeId), 10)][ENNAME]
		}
	} else if record.Head.Config.Meta.RoomId != 0 { //friendly
		lobby = ": " + strconv.FormatUint(uint64(record.Head.Config.Meta.RoomId), 10) //can set room number as lobby number
		ruledisp += RUNES["friendly"][NAMEPREF]                                       //"Friendly";
		nakas = record.Head.Config.Mode.DetailRule.DoraCount
		if 3 == nplayers {
			TSUMOLOSSOFF = !record.Head.Config.Mode.DetailRule.HaveZimosun
		} else {
			TSUMOLOSSOFF = false
		}
	} else if record.Head.Config.Meta.ContestUid != 0 { //tourney
		lobby = ": " + strconv.FormatUint(uint64(record.Head.Config.Meta.ContestUid), 10)
		ruledisp += RUNES["tournament"][NAMEPREF] //"Tournament";
		nakas = record.Head.Config.Mode.DetailRule.DoraCount
		if 3 == nplayers {
			TSUMOLOSSOFF = !record.Head.Config.Mode.DetailRule.HaveZimosun
		} else {
			TSUMOLOSSOFF = false
		}
	}
	if 1 == record.Head.Config.Mode.Mode {
		ruledisp += RUNES["tonpuu"][NAMEPREF]
	} else if 2 == record.Head.Config.Mode.Mode {
		ruledisp += RUNES["hanchan"][NAMEPREF]
	}
	if record.Head.Config.Meta.ModeId == 0 && record.Head.Config.Mode.DetailRule.DoraCount == 0 {
		if JPNAME != NAMEPREF {
			ruledisp += RUNES["nored"][NAMEPREF]
		}
		res["rule"] = map[string]interface{}{"disp": ruledisp, "aka53": 0, "aka52": 0, "aka51": 0}
	} else {
		if JPNAME == NAMEPREF {
			ruledisp += RUNES["red"][JPNAME]
		}
		aka52, aka51 := 0, 0
		if 4 == nakas {
			aka52 = 2
		} else {
			aka52 = 1
		}

		if 4 == nplayers {
			aka51 = 1
		} else {
			aka51 = 0
		}

		res["rule"] = map[string]interface{}{"disp": ruledisp, "aka53": 1, "aka52": aka52, "aka51": aka51}
	}

	res["lobby"] = 0 //tenhou custom lobby - could be tourney id or friendly room for mjs. appending to title instead to avoid 3->C etc. in tenhou.net/5
	// autism to fix logs with AI
	// ranks
	res["dan"] = []string{"", "", "", ""}
	for _, e := range record.Head.Accounts {
		if JPNAME == NAMEPREF {
			res["dan"].([]string)[e.Seat] = CFG_LEVEL[strconv.FormatUint(uint64(e.Level.Id), 10)][JPNAME]
		} else {
			res["dan"].([]string)[e.Seat] = CFG_LEVEL[strconv.FormatUint(uint64(e.Level.Id), 10)][ENNAME]
		}
	} // 段位罢了。例如：雀杰一星、雀杰三星。自己实现一个e.level.id -> full_name_en的映射（map）就行了。

	// level score, no real analog to rate
	res["rate"] = []uint32{0, 0, 0, 0}
	for _, e := range record.Head.Accounts {
		res["rate"].([]uint32)[e.Seat] = e.Level.Score //level score, closest thing to rate
	}
	// sex
	res["sx"] = []string{"C", "C", "C", "C"}
	for _, e := range record.Head.Accounts {
		sex := CFG_SEX[strconv.FormatUint(uint64(e.Character.Charid), 10)][JPNAME]
		if sex == "1" {
			res["sx"].([]string)[e.Seat] = "F"
		} else if sex == "2" {
			res["sx"].([]string)[e.Seat] = "M"
		} else {
			res["sx"].([]string)[e.Seat] = "C"
		}
	}
	// >names
	res["name"] = []string{"AI", "AI", "AI", "AI"}
	for _, e := range record.Head.Accounts {
		res["name"].([]string)[e.Seat] = e.Nickname
	}
	// clean up for sanma AI
	if 3 == nplayers {
		res["name"].([]string)[3] = ""
		res["sx"].([]string)[3] = ""
	}
	// scores
	var scores [4][3]int32
	for i, e := range record.Head.Result.Players {
		scores[i][0] = int32(e.Seat)
		scores[i][1] = e.PartPoint_1
		scores[i][2] = e.TotalPoint / 1000
	}
	res["sc"] = []int32{0, 0, 0, 0, 0, 0, 0, 0}
	for _, e := range scores {
		res["sc"].([]int32)[2*e[0]] = e[1]
		res["sc"].([]int32)[2*e[0]+1] = e[2]
	}
	// optional title - why not give the room and put the timestamp here; 1000 for unix to .js timestamp convention
	tm := time.Unix(int64(record.Head.EndTime), 0)
	res["title"] = [2]string{ruledisp + lobby, tm.Format("2006/01/02 15:04:05")}
	// optionally dump mjs records NOTE: this will likely make the file too large for tenhou.net/5 viewer
	if VERBOSELOG {
		res["mjshead"] = record.Head
		res["mjslog"] = mjslog
		type_mjslog := []string{}
		for _, e := range mjslog {
			type_mjslog = append(type_mjslog, fmt.Sprintf("%T", e))
		}
		res["mjsrecordtypes"] = type_mjslog
	}
	return res
}

func Downloadlog(record *message.ResGameRecord) []byte {
	results := parse(record)
	if PRETTY {
		jsonData, err := json.MarshalIndent(results, "", "    ")
		if err != nil {
			return nil
		}
		return jsonData
	} else {
		jsonData, err := json.Marshal(results)
		if err != nil {
			return nil
		}
		return jsonData
	}
}
