package protocol

import (
	"TeenPatti/TPABGameServer/pkg/constant"
)

type ReJoinDeskRequest struct {
	DeskNo string `json:"deskId"`
}

type ReJoinDeskResponse struct {
	Code  int    `json:"code"`
	Error string `json:"error"`
}


type ReEnterDeskRequest struct {
	DeskNo string `json:"deskId"`
}

type ReEnterDeskResponse struct {
	Code  int    `json:"code"`
	Error string `json:"error"`
}

//---------------------------------------------------------------------------------------------------
type TableInfo struct {
	DeskNo    string              `json:"deskId"`
	CreatedAt int64               `json:"createdAt"`
	Creator   int64               `json:"creator"`
	Title     string              `json:"title"`
	Desc      string              `json:"desc"`



	Status    constant.DeskStatus `json:"status"`
	Round     uint32              `json:"round"`
	Mode      int                 `json:"mode"`
}



type JoinDeskRequest struct {

	Version string `json:"version"`

	//AccountId int64         `json:"acId"`
	//DeskNo string `json:"deskId"`

}

type JoinDeskResponse struct {
	Code      int       `json:"code"`
	Error     string    `json:"error"`

	TableInfo TableInfo `json:"tableInfo"`
}

type DestoryDeskRequest struct {
	DeskNo string `json:"deskId"`
}

type PlayerOfflineStatus struct {
	Uid     int64 `json:"uid"`
	Offline bool  `json:"offline"`
}

//-----------------------------------------------------------------------------------------------