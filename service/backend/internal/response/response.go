package response

type Body struct {
	Msg  string `json:"msg"`
	Code int    `json:"code"`
	Data any    `json:"data,omitempty"`
}

func Success(msg string, code int) Body {
	return Body{
		Msg:  msg,
		Code: code,
	}
}

func SuccessData(msg string, code int, data any) Body {
	return Body{
		Msg:  msg,
		Code: code,
		Data: data,
	}
}

func Fail(msg string, code int) Body {
	return Body{
		Msg:  msg,
		Code: code,
	}
}

func FailData(msg string, code int, data any) Body {
	return Body{
		Msg:  msg,
		Code: code,
		Data: data,
	}
}

func Error(msg string, code int) Body {
	return Body{
		Msg:  msg,
		Code: code,
	}
}

func ErrorData(msg string, code int, data any) Body {
	return Body{
		Msg:  msg,
		Code: code,
		Data: data,
	}
}
