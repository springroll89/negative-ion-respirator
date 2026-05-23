package handler

import (
	"github.com/gin-gonic/gin"
	"negative-ion-respirator/backend/internal/service"
	"strconv"
)

type BatchHandler struct {
	batchSvc  *service.BatchService
	reportSvc *service.ReportService
}

func NewBatchHandler(batchSvc *service.BatchService, reportSvc *service.ReportService) *BatchHandler {
	return &BatchHandler{batchSvc: batchSvc, reportSvc: reportSvc}
}

func (h *BatchHandler) CreateBatchConfig(c *gin.Context) {
	var req service.BatchConfigReq
	if err := c.ShouldBindJSON(&req); err != nil {
		Err(c, 40001, 400, "invalid request: "+err.Error())
		return
	}
	task, err := h.batchSvc.CreateBatchConfig(c.Request.Context(), req)
	if err != nil {
		Err(c, 50001, 500, err.Error())
		return
	}
	OK(c, task)
}

func (h *BatchHandler) GetTaskStatus(c *gin.Context) {
	taskID, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	task, err := h.batchSvc.GetTask(taskID)
	if err != nil {
		Err(c, 40001, 404, err.Error())
		return
	}
	OK(c, task)
}

func (h *BatchHandler) GetDashboard(c *gin.Context) {
	stats, err := h.reportSvc.GetDashboard(c.Request.Context())
	if err != nil {
		Err(c, 50001, 500, err.Error())
		return
	}
	OK(c, stats)
}

func (h *BatchHandler) GetReport(c *gin.Context) {
	var req service.UsageReportRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		Err(c, 40001, 400, "invalid request")
		return
	}
	report, err := h.reportSvc.GetFullReport(c.Request.Context(), req)
	if err != nil {
		Err(c, 50001, 500, err.Error())
		return
	}
	OK(c, report)
}
