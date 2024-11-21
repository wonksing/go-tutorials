package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type userHandler struct {
}

func NewUserHandler() *userHandler {
	return &userHandler{}
}

func (h *userHandler) GetUser(c *gin.Context) {
	userID := c.Param("userId")
	c.JSON(http.StatusOK, &GetUserRes{
		ID:   userID,
		Name: "wonk",
	})
}

type GetUserRes struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}
