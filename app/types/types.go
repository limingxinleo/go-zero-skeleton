// Code generated by goctl. DO NOT EDIT.
// goctl 1.7.3

package types

type FromRequest struct {
	Name string `form:"name,optional,default=world"`
}

type Response[T any] struct {
	Code    int    `json:"code"`
	Data    T      `json:"data"`
	Message string `json:"message"`
	TraceId string `json:"trace_id"`
}
