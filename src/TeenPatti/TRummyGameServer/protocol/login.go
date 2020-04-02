package protocol

//=================================================================================================================

//玩家登陆时发送的消息
type LoginToGameServerRequest struct {
	Version string `json:"version"` //客户端版本号
	Token   string `json:"token"`   //上传的token
	Uid     int64  `json:"uid"`     //这个Uid是大厅服务器传给玩家的

}

//服务器回复给玩家的消息
type LoginToGameServerResponse struct {
	Success bool  `json:"success"` //是否成功
	Uid     int64 `json:"uId"`     //帐号的ID

	Nickname string `json:"nickname"` //玩家名秒
	HeadUrl  string `json:"headURL"`
	Sex      int    `json:"sex"`

	Error string `json:"error"` //错别原因

}
