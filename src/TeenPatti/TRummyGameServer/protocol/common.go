package protocol

//大厅用户数据
type FaceBookGetPlayerMsg struct {
	Avatar      string  `json:"avatar"`
	Chips       float64 `json:"chips"`
	Diamond     int32   `json:"diamond"`
	ID          int32   `json:"id"`
	Integral    int32   `json:"integral"`
	LastLoginAt int32   `json:"last_login_at"`
	Level       int32   `json:"level"`
	Name        string  `json:"name"`
	RegisterAt  int32   `json:"register_at"`
	Role        int8    `json:"role"`
	Sex         int8    `json:"sex"`
}

//获取facebook的用户信息
type FaceBookGetPlayerMsgData struct {
	Code int32                `json:"code"`
	Data FaceBookGetPlayerMsg `json:"data"`
}
