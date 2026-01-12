package api

import (
	"testing"

	"github.com/PRPO-skupina-02/common/validation"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func TestingRouter(t *testing.T, db *gorm.DB) *gin.Engine {
	router := gin.Default()
	trans, err := validation.RegisterValidation()
	require.NoError(t, err)
	Register(router, db, trans)

	return router
}
