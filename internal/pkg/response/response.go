package response

type Response struct {
	Code int         `json:"code"`
	Msg  string      `json:"msg"`
	Data interface{} `json:"data,omitempty"`
}

const (
	CodeSuccess = 0

	CodeParamError   = 40001
	CodeUnauthorized = 40002
	CodeForbidden    = 40003
	CodeNotFound     = 40004

	CodeServerError = 50001
	CodeDBError     = 50002

	CodeUserExist      = 10001
	CodeUserNotFound   = 10002
	CodePasswordWrong  = 10003
	CodeTokenInvalid   = 10004
	CodeTokenExpired   = 10005
	CodeTokenMalformed = 10006

	CodeSlotFull       = 20001
	CodeSlotNotFound   = 20002
	CodeSlotPast       = 20003
	CodeApplianceInvalid = 20004
	CodeAppointmentNotFound = 20005
	CodeAppointmentDone = 20006
	CodeAppointmentCanceled = 20007
	CodeStatusInvalid   = 20008
	CodeImageTooMany    = 20009
	CodeImageTooFew     = 20010
	CodeImageInvalidExt = 20011
	CodeImageTooLarge   = 20012
	CodeImageSaveFailed = 20013
)

var msgMap = map[int]string{
	CodeSuccess: "success",

	CodeParamError:   "参数错误",
	CodeUnauthorized: "未授权",
	CodeForbidden:    "无权限访问",
	CodeNotFound:     "资源不存在",

	CodeServerError: "服务器内部错误",
	CodeDBError:     "数据库操作错误",

	CodeUserExist:      "用户已存在",
	CodeUserNotFound:   "用户不存在",
	CodePasswordWrong:  "密码错误",
	CodeTokenInvalid:   "Token无效",
	CodeTokenExpired:   "Token已过期",
	CodeTokenMalformed: "Token格式错误",

	CodeSlotFull:            "该时段已满",
	CodeSlotNotFound:        "时段不存在",
	CodeSlotPast:            "该时段已过期",
	CodeApplianceInvalid:    "家电类型无效",
	CodeAppointmentNotFound: "预约不存在",
	CodeAppointmentDone:     "预约已完成，无法修改",
	CodeAppointmentCanceled: "预约已取消，无法修改",
	CodeStatusInvalid:       "状态值无效",
	CodeImageTooMany:        "图片数量过多，最多5张",
	CodeImageTooFew:         "图片数量不足，至少3张",
	CodeImageInvalidExt:     "图片格式不支持，仅允许jpg/jpeg/png/webp",
	CodeImageTooLarge:       "图片过大，单张不超过5MB",
	CodeImageSaveFailed:     "图片保存失败",
}

func GetMsg(code int) string {
	if msg, ok := msgMap[code]; ok {
		return msg
	}
	return "未知错误"
}

func Success(data interface{}) Response {
	return Response{
		Code: CodeSuccess,
		Msg:  GetMsg(CodeSuccess),
		Data: data,
	}
}

func Fail(code int) Response {
	return Response{
		Code: code,
		Msg:  GetMsg(code),
	}
}

func FailWithMsg(code int, msg string) Response {
	return Response{
		Code: code,
		Msg:  msg,
	}
}
