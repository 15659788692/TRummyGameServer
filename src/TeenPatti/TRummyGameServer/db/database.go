package db

import "database/sql"

//import (
//	"database/sql"
//	_ "github.com/go-sql-driver/mysql"
//)

var DBCon *sql.DB

//func init() {
//	userName := conf.Conf.DB.USERNAME
//	pwd := conf.Conf.DB.USERNAME
//	net := conf.Conf.DB.NETWORK
//	host := conf.Conf.DB.HOST
//	port := conf.Conf.DB.PORT
//	dbname := conf.Conf.DB.DBNAME
//
//	dsn := fmt.Sprintf("%s:%s@%s(%s:%d)/%s", userName, pwd, net, host, port, dbname)
//	DB, err := sql.Open("mysql", dsn)
//	if err != nil {
//		panic(err)
//	}
//
//	DB.SetConnMaxLifetime(100 * time.Second)
//	DB.SetMaxOpenConns(100)
//	DB.SetMaxIdleConns(16)
//	DBCon = DB
//
//	if err1 := DB.Ping(); err1 != nil {
//		fmt.Println(err1)
//		panic(err1)
//	}
//
//	fmt.Println("数据库连接成功 ")
//}

///*
//Do:更新用户金币
//Author:信。
//Date:2020/03/12
//Modify:nil
//*/
//func UpdatePlayerGold(gold float64, id int32) {
//	var result sql.Result
//	var err error
//	result, err = DBCon.Exec("UPDATE user set chips=chips+? where id=?", gold, id)
//
//	if err != nil {
//		str := fmt.Sprintf("UpdatePlayerGold failed,err:%v ,uid %v ,gold %v", err, id, gold)
//		logInit.AddLog(str)
//		return
//	}
//
//	_, err = result.RowsAffected()
//	if err != nil {
//		str := fmt.Sprintf("UpdatePlayerGold failed,err:%v ,uid %v ,gold %v", err, id, gold)
//		logInit.AddLog(str)
//		return
//	}
//}
//
///*
//Do:添加玩家每次投注记录
//Author:信。
//Date:2020/03/12
//Modify:nil
//*/
//func AddPlayerBetRecord(uid int32, userName string, recommenderID int32, gameType int32, roundCount int32, loseWin float64, invalid int32, seatID int8, minBet float64, deskID string, beforeChips float64, betChips float64, region int8) {
//	stmt, err := DBCon.Prepare("INSERT bet (uid,`name`,recommender_id,`type`,round_count,status,invalid,seat,bottom,desk_id,before_chips,bet_chips,region,created_at) values (?,?,?,?,?,?,?,?,?,?,?,?,?,?)")
//	if err != nil {
//		str := fmt.Sprintf("AddPlayerBetRecord failed,err:%v,%v", err, err)
//		logInit.AddLog(str)
//		return
//	}
//
//	res, err := stmt.Exec(uid, userName, recommenderID, gameType, roundCount, loseWin, invalid, seatID, minBet, deskID, beforeChips, betChips, region, time.Now().Unix())
//	if err != nil {
//		str := fmt.Sprintf("AddPlayerBetRecord failed,err:%v,%v", err, err)
//		logInit.AddLog(str)
//		return
//	}
//
//	_, err = res.LastInsertId()
//	if err != nil {
//		str := fmt.Sprintf("AddPlayerBetRecord failed,err:%v,%v", err, err)
//		logInit.AddLog(str)
//		return
//	}
//
//}
//
///*
//Do:添加玩家每一局结算
//Author:信。
//Date:2020/03/12
//Modify:nil
//*/
//func AddPlayerRoundRecord(uid int32, name string, recommender_id int32, gameType int32, desk_id string, bottom float64, bring float64, lose_win float64, seat int8, pumped float64, pumped_scale float64, bet_chips float64, after_chips float64, round_count int32) {
//	stmt, err := DBCon.Prepare("INSERT round (uid, `name`, recommender_id, `type`, desk_id, bottom, bring, lose_win,seat,pumped,pumped_scale,bet_chips,after_chips,round_count,created_at) values (?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)")
//	if err != nil {
//		str := fmt.Sprintf("AddPlayerRoundRecord failed,err:%v ,uid  %v", err, uid)
//		logInit.AddLog(str)
//		return
//	}
//
//	res, err := stmt.Exec(uid, name, recommender_id, gameType, desk_id, bottom, bring, lose_win, seat, pumped, pumped_scale, bet_chips, after_chips, round_count, time.Now().Unix())
//	if err != nil {
//		str := fmt.Sprintf("AddPlayerRoundRecord failed,err:%v ,uid  %v", err, uid)
//		logInit.AddLog(str)
//		return
//	}
//
//	_, err = res.LastInsertId()
//	if err != nil {
//		str := fmt.Sprintf("AddPlayerRoundRecord failed,err:%v ,uid  %v", err, uid)
//		logInit.AddLog(str)
//		return
//	}
//}
//
///*
//Do:桌子解散添加记录
//Author:信。
//Date:2020/03/12
//Modify:nil
//*/
//func AddDeskRoundRecord(desk_id string, gameType int32, status int32, people int32, bring float64, win float64, lose float64, bottom float64, pumped float64, duration int64, round_count int32, chips float64) {
//	stmt, err := DBCon.Prepare("INSERT desk (desk_id,`type`,status,people,bring,win,lose,bottom,pumped,duration,round_count,chips,created_at) values (?,?,?,?,?,?,?,?,?,?,?,?,?)")
//	if err != nil {
//		str := fmt.Sprintf("AddDeskRoundRecord failed,err:%v ,desk_id  %v", err, desk_id)
//		logInit.AddLog(str)
//		return
//	}
//
//	res, err := stmt.Exec(desk_id, gameType, status, people, bring, win, lose, bottom, pumped, duration, round_count, chips, time.Now().Unix())
//	if err != nil {
//		str := fmt.Sprintf("AddDeskRoundRecord failed,err:%v ,desk_id  %v", err, desk_id)
//		logInit.AddLog(str)
//		return
//	}
//
//	_, err = res.LastInsertId()
//	if err != nil {
//		str := fmt.Sprintf("AddDeskRoundRecord failed,err:%v ,desk_id  %v", err, desk_id)
//		logInit.AddLog(str)
//		return
//	}
//}
