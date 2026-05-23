package service

import (
	"context"
	"negative-ion-respirator/backend/internal/repository"
	"time"
)

type ReportService struct {
	deviceRepo    *repository.DeviceRepo
	orderRepo     *repository.OrderRepo
	telemetryRepo *repository.TelemetryRepo
}

func NewReportService(deviceRepo *repository.DeviceRepo, orderRepo *repository.OrderRepo, telemetryRepo *repository.TelemetryRepo) *ReportService {
	return &ReportService{
		deviceRepo:    deviceRepo,
		orderRepo:     orderRepo,
		telemetryRepo: telemetryRepo,
	}
}

type DashboardStats struct {
	TotalDevices  int   `json:"total_devices"`
	OnlineDevices int   `json:"online_devices"`
	TodayOrders   int   `json:"today_orders"`
	TodayRevenue  int64 `json:"today_revenue"`
	ActiveOrders  int   `json:"active_orders"`
}

func (s *ReportService) GetDashboard(ctx context.Context) (*DashboardStats, error) {
	devices, total, err := s.deviceRepo.List(ctx, 0, 10000)
	if err != nil {
		return nil, err
	}

	stats := &DashboardStats{TotalDevices: total}
	for _, d := range devices {
		if d.Status == "online" || d.Status == "running" {
			stats.OnlineDevices++
		}
	}

	// Today's orders would come from order repository
	// For now, placeholder values
	stats.TodayOrders = 0
	stats.TodayRevenue = 0
	stats.ActiveOrders = 0

	return stats, nil
}

type DeviceReport struct {
	DeviceID      int64   `json:"device_id"`
	DeviceSN      string  `json:"device_sn"`
	TotalSessions int     `json:"total_sessions"`
	TotalDuration int     `json:"total_duration"`
	TotalRevenue  int64   `json:"total_revenue"`
	AvgOutTemp    float64 `json:"avg_out_temp"`
	FaultCount    int     `json:"fault_count"`
}

func (s *ReportService) GetDeviceUsageReport(ctx context.Context, start, end time.Time) ([]DeviceReport, error) {
	// Query device logs for the time range
	devices, _, err := s.deviceRepo.List(ctx, 0, 10000)
	if err != nil {
		return nil, err
	}

	var reports []DeviceReport
	for _, d := range devices {
		logs, err := s.telemetryRepo.QueryByDevice(ctx, d.ID, start, end, 10000)
		if err != nil {
			continue
		}

		report := DeviceReport{
			DeviceID: d.ID,
			DeviceSN: d.DeviceSN,
		}

		var tempSum float64
		var tempCount int
		for _, entry := range logs {
			if entry.EventType != "" {
				report.FaultCount++
			}
			if entry.Status == "running" && entry.OutTemp > 0 {
				tempSum += entry.OutTemp
				tempCount++
			}
		}
		if tempCount > 0 {
			report.AvgOutTemp = tempSum / float64(tempCount)
		}
		report.TotalSessions = len(logs) / 12 // ~1 session per 12 status reports (60s)

		reports = append(reports, report)
	}

	return reports, nil
}

type RevenueReport struct {
	Period       string `json:"period"`
	TotalOrders  int    `json:"total_orders"`
	TotalRevenue int64  `json:"total_revenue"`
	AvgDuration  int    `json:"avg_duration_sec"`
	TopDeviceID  int64  `json:"top_device_id"`
	TopDeviceSN  string `json:"top_device_sn"`
}

func (s *ReportService) GetRevenueReport(ctx context.Context, period string) (*RevenueReport, error) {
	// Production would compute from orders table with time-based aggregation
	return &RevenueReport{
		Period:       period,
		TotalOrders:  0,
		TotalRevenue: 0,
		AvgDuration:  0,
	}, nil
}

// UsageReportRequest represents a report query
type UsageReportRequest struct {
	StartDate string `json:"start_date" form:"start_date"`
	EndDate   string `json:"end_date" form:"end_date"`
	Period    string `json:"period" form:"period"` // daily, weekly, monthly
}

// UsageReportResponse wraps report data
type UsageReportResponse struct {
	DeviceReports []DeviceReport  `json:"device_reports"`
	Revenue       *RevenueReport  `json:"revenue"`
	Dashboard     *DashboardStats `json:"dashboard"`
}

func (s *ReportService) GetFullReport(ctx context.Context, req UsageReportRequest) (*UsageReportResponse, error) {
	start, _ := time.Parse("2006-01-02", req.StartDate)
	end, _ := time.Parse("2006-01-02", req.EndDate)
	if req.StartDate == "" {
		start = time.Now().AddDate(0, 0, -7)
		end = time.Now()
	}

	deviceReports, _ := s.GetDeviceUsageReport(ctx, start, end)
	revenue, _ := s.GetRevenueReport(ctx, req.Period)
	dashboard, _ := s.GetDashboard(ctx)

	return &UsageReportResponse{
		DeviceReports: deviceReports,
		Revenue:       revenue,
		Dashboard:     dashboard,
	}, nil
}
