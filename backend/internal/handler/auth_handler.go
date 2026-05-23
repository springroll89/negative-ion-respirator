package handler

import (
	"github.com/gin-gonic/gin"
	"negative-ion-respirator/backend/internal/model"
	"negative-ion-respirator/backend/internal/service"
	"strings"
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

func (h *AuthHandler) Refresh(c *gin.Context) {
	auth := c.GetHeader("Authorization")
	if auth == "" || !strings.HasPrefix(auth, "Bearer ") {
		Err(c, 40101, 401, "missing token")
		return
	}

	claims, err := h.svc.ValidateToken(strings.TrimPrefix(auth, "Bearer "))
	if err != nil {
		Err(c, 40102, 401, "invalid or expired token")
		return
	}

	username, _ := claims["username"].(string)
	resp, err := h.svc.RefreshToken(c.Request.Context(), username)
	if err != nil {
		Err(c, 40102, 401, "failed to refresh token")
		return
	}
	OK(c, resp)
}
