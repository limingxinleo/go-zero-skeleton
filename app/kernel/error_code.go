package kernel

type ErrorCodeInterface interface {
	GetCode() int
	GetMessage() string
	Error() string
}
