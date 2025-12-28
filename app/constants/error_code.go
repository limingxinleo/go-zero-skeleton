package constants

type ErrorCode struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	err     error
}

func (e *ErrorCode) Error() string {
	return e.Message
}

func (e *ErrorCode) GetMessage() string {
	return e.Message
}

func (e *ErrorCode) GetCode() int {
	return e.Code
}

func (e *ErrorCode) Err() error {
	return e.err
}

func (e *ErrorCode) WithError(err error) *ErrorCode {
	e.err = err
	return e
}

func (e *ErrorCode) WithMessage(message string) *ErrorCode {
	e.Message = message
	return e
}

var ServerError = &ErrorCode{Code: 500, Message: "服务器内部错误"}
