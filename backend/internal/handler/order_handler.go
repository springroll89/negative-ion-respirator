package handler

import (
	"github.com/gin-gonic/gin"
	"negative-ion-respirator/backend/internal/model"
	"negative-ion-respirator/backend/internal/service"
)

type OrderHandler struct{ svc *service.OrderService }

func NewOrderHandler(svc *service.OrderService) *OrderHandler { return &OrderHandler{svc: svc} }

func (h *OrderHandler) Create(c *gin.Context) {
	var req model.CreateOrderReq
	if err := c.ShouldBindJSON(&req); err != nil {
		Err(c, 40001, 400, "invalid request: "+err.Error())
		return
	}
	o, err := h.svc.Create(c.Request.Context(), req)
	if err != nil {
		Err(c, 50001, 500, err.Error())
		return
	}
	OK(c, o)
}

func (h *OrderHandler) Start(c *gin.Context) {
	var req struct {
		TID string `json:"tid" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		Err(c, 40001, 400, "invalid request")
		return
	}
	if err := h.svc.Start(c.Request.Context(), req.TID); err != nil {
		Err(c, 50001, 500, err.Error())
		return
	}
	OK(c, gin.H{"status": "started"})
}

func (h *OrderHandler) Stop(c *gin.Context) {
	var req struct {
		TID string `json:"tid" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		Err(c, 40001, 400, "invalid request")
		return
	}
	o, err := h.svc.Stop(c.Request.Context(), req.TID)
	if err != nil {
		Err(c, 50001, 500, err.Error())
		return
	}
	OK(c, o)
}

func (h *OrderHandler) Query(c *gin.Context) {
	tid := c.Query("tid")
	o, err := h.svc.FindByTID(c.Request.Context(), tid)
	if err != nil {
		Err(c, 40004, 404, "order not found")
		return
	}
	OK(c, o)
}
