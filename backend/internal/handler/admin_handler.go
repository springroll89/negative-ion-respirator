package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"negative-ion-respirator/backend/internal/service"
)

type AdminHandler struct {
	deviceSvc *service.DeviceService
}

func NewAdminHandler(deviceSvc *service.DeviceService) *AdminHandler {
	return &AdminHandler{deviceSvc: deviceSvc}
}

func (h *AdminHandler) UpdateDeviceConfig(c *gin.Context) {
	var req struct {
		DeviceID      int64 `json:"device_id" binding:"required"`
		MaxHeatTemp   int   `json:"max_heat_temp" binding:"required"`
		TargetOutTemp int   `json:"target_out_temp" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		Err(c, 40001, 400, "invalid request")
		return
	}
	if err := h.deviceSvc.UpdateConfig(c.Request.Context(), req.DeviceID, req.MaxHeatTemp, req.TargetOutTemp); err != nil {
		Err(c, 50001, 500, err.Error())
		return
	}
	OK(c, gin.H{"status": "config_updated"})
}

func (h *AdminHandler) GetDeviceStatus(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	d, err := h.deviceSvc.GetDevice(c.Request.Context(), id)
	if err != nil {
		Err(c, 40001, 404, "device not found")
		return
	}
	OK(c, d)
}
