package service

import (
	"context"
	"github.com/stretchr/testify/assert"
	"main/app"
	"main/app/types"
	"testing"
)

func TestIndexService_Index(t *testing.T) {
	service := NewIndexService(context.TODO(), app.ServiceContext)

	result, _ := service.Index(&types.FromRequest{
		Name: "limingxinleo",
	})

	assert.Equal(t, "Hi limingxinleo, welcome to main-api", result)
}
