package http

type ErrorCodeInterface interface {
	GetCode() int
	GetMessage() string
}
