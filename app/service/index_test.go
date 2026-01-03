package service

import (
	"context"
	"testing"

	"github.com/limingxinleo/go-zero-skeleton/app"
	"github.com/limingxinleo/go-zero-skeleton/app/types"
	"github.com/stretchr/testify/assert"
)

func TestIndexService_Index(t *testing.T) {
	service := NewIndexService(context.TODO(), app.GetApplication().ServiceContext)

	result, _ := service.Index(&types.FromRequest{
		Name: "limingxinleo",
	})

	assert.Equal(t, "Hi limingxinleo, welcome to main-api", result)
}
