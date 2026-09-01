package service

import (
	"context"
	"testing"

	"github.com/limingxinleo/go-zero-skeleton/app"
	"github.com/limingxinleo/go-zero-skeleton/app/types"
	"github.com/stretchr/testify/assert"
)

func TestIndexService_Index(t *testing.T) {
	s := NewIndexService(context.Background(), app.GetApplication().ServiceContext)

	result, err := s.Index(&types.FromRequest{
		Name: "limingxinleo",
	})

	assert.NoError(t, err)
	assert.Equal(t, "Hi limingxinleo, welcome to main-api", result)
}
