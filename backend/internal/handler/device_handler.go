package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"negative-ion-respirator/backend/internal/service"
)

type DeviceHandler struct{ svc *service.DeviceService }

func NewDeviceHandler(svc *service.DeviceService) *DeviceHandler { return &DeviceHandler{svc: svc} }

func (h *DeviceHandler) GetDevice(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	d, err := h.svc.GetDevice(c.Request.Context(), id)
	if err != nil {
		Err(c, 40001, 404, "device not found")
		return
	}
	OK(c, d)
}

func (h *DeviceHandler) ListDevices(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	devices, total, err := h.svc.ListDevices(c.Request.Context(), page, pageSize)
	if err != nil {
		Err(c, 50001, 500, err.Error())
		return
	}
	OKWithMeta(c, devices, gin.H{"total": total, "page": page, "page_size": pageSize})
}

func (h *DeviceHandler) Register(c *gin.Context) {
	var req struct {
		DeviceSN   string `json:"device_sn" binding:"required"`
		DeviceName string `json:"device_name"`
		RegionCode string `json:"region_code"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		Err(c, 40001, 400, "invalid request: "+err.Error())
		return
	}
	d, err := h.svc.Register(c.Request.Context(), req.DeviceSN, req.DeviceName, req.RegionCode)
	if err != nil {
		Err(c, 50001, 500, err.Error())
		return
	}
	OK(c, d)
}
