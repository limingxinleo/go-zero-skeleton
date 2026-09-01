package kernel

// ErrorCode 统一错误码契约：业务错误实现该接口，由 Send 统一输出 Code/Message。
// 它本身是标准 error（含 Unwrap/Is），可直接配合 errors.Is / fmt.Errorf("%w") 等标准工具链使用。
type ErrorCode interface {
	Code() int
	Message() string
	error
	Unwrap() error
}
