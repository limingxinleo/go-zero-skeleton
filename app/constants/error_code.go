package constants

type ErrorCode struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

var ServerError = &ErrorCode{Code: 500, Message: "服务器内部错误"}

func (e *ErrorCode) Error() string {
	return e.Message
}

func (e *ErrorCode) GetMessage() string {
	return e.Message
}

func (e *ErrorCode) GetCode() int {
	return e.Code
}

func (e *ErrorCode) WithMessage(message string) *ErrorCode {
	e.Message = message
	return e
}
