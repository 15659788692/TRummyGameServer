package protocol

//=================================================================================================================

//玩家登陆时发送的消息
type LoginToGameServerRequest struct {

	Uid     int64  `json:"uid"`
	Name    string `json:"name"`
	HeadUrl string `json:"headUrl"`
	Sex     int    `json:"sex"`        //[0]未知 [1]男 [2]女
	FangKa  int    `json:"fangka"`
	IP      string `json:"ip"`

}

//服务器回复给玩家的消息
type LoginToGameServerResponse struct {
	Uid      int64  `json:"acId"`
	Nickname string `json:"nickname"`
	HeadUrl  string `json:"headURL"`
	Sex      int    `json:"sex"`
	FangKa   int    `json:"fangka"`
}

