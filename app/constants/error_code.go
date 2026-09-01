package constants

// ErrorCode 业务错误码。通过 NewErrorCode 构造，按业务模块集中定义为包级变量：
//
//	var XxxError = NewErrorCode(1001, "描述")
//
// 错误码是全局单例，WithError / WithMessage 均返回新副本而非原地修改，
// 保证并发请求间互不覆盖（详见各方法注释）。
type ErrorCode struct {
	code    int
	message string
	err     error
}

func NewErrorCode(code int, message string) *ErrorCode {
	return &ErrorCode{code: code, message: message}
}

func (e *ErrorCode) Error() string {
	return e.message
}

func (e *ErrorCode) Code() int {
	return e.code
}

func (e *ErrorCode) Message() string {
	return e.message
}

// Unwrap 返回底层错误（仅写入日志，不暴露给客户端），兼容 errors.Is / errors.As。
func (e *ErrorCode) Unwrap() error {
	return e.err
}

// Is 让 errors.Is 按错误码匹配：WithError / WithMessage 产生的副本仍识别为同一错误码。
func (e *ErrorCode) Is(target error) bool {
	t, ok := target.(*ErrorCode)
	return ok && t.code == e.code
}

// WithError 返回附加了底层错误的新副本，原错误码保持不变（并发安全）。
// 底层错误仅写入日志，不改变对外的 Code 与 Message。
func (e *ErrorCode) WithError(err error) *ErrorCode {
	return &ErrorCode{code: e.code, message: e.message, err: err}
}

// WithMessage 返回覆盖了对外提示的新副本，原错误码保持不变（并发安全）。
func (e *ErrorCode) WithMessage(message string) *ErrorCode {
	return &ErrorCode{code: e.code, message: message, err: e.err}
}

var ServerError = NewErrorCode(500, "服务器内部错误")
