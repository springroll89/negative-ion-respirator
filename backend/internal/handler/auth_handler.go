package handler

import (
	"github.com/gin-gonic/gin"
	"negative-ion-respirator/backend/internal/model"
	"negative-ion-respirator/backend/internal/service"
)

type AuthHandler struct{ svc *service.AuthService }

func NewAuthHandler(svc *service.AuthService) *AuthHandler { return &AuthHandler{svc: svc} }

func (h *AuthHandler) Login(c *gin.Context) {
	var req model.LoginReq
	if err := c.ShouldBindJSON(&req); err != nil {
		Err(c, 40001, 400, "invalid request")
		return
	}
	resp, err := h.svc.Login(c.Request.Context(), req)
	if err != nil {
		Err(c, 40102, 401, "invalid credentials")
		return
	}
	OK(c, resp)
}
