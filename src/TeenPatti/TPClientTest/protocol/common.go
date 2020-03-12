package protocol

type AppInfo struct {
	Name            string                       `json:"name"`             //应用名
	AppID           string                       `json:"appid"`            //应用id
	AppKey          string                       `json:"appkey"`           //应用key
	RedirectURI     string                       `json:"redirect_uri"`     //注册时填的redirect_uri
	Extra           string                       `json:"extra"`            //额外信息
	ThirdProperties map[string]map[string]string `json:"third_properties"` //此app在第三方平台(eg: wechat)上的相关配置
}

type UserInfo struct {
	UID         int64  `json:"uid"`
	Name        string `json:"name"`            //用户名, 可空,当非游客注册时用户名与手机号必须至少出现一项
	Phone       string `json:"phone"`           //手机号,可空
	Role        int    `json:"role"`            //玩家类型
	Status      int    `json:"status"`          //状态
	IsOnline    int    `json:"is_online"`       //是否在线
	LastLoginAt int64  `json:"last_login_time"` //最后登录时间
}

type Device struct {
	IMEI   string `json:"imei"`   //设备的imei号
	OS     string `json:"os"`     //os版本号
	Model  string `json:"model"`  //硬件型号
	IP     string `json:"ip"`     //内网IP
	Remote string `json:"remote"` //外网IP
}

type StringResponse struct {
	Code int    `json:"code"` //状态码
	Data string `json:"data"` //字符串数据
}

type CommonResponse struct {
	Code int         `json:"code"` //状态码
	Data interface{} `json:"data"` //整数状态
}

var SuccessResponse = StringResponse{0, "success"}

var EmptyMessage = &None{}

type EmptyRequest struct{}

var SuccessMessage = &StringMessage{Message: "success"}

type None struct{}

type StringMessage struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type ErrorResponse struct {
	Code  int    `json:"code"`
	Error string `json:"error"`
}

type ErrorMessage struct {
	ErrorType int    `json:"errorType"`
	Message   string `json:"msg"`
}


