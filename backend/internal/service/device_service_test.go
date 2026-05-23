package service_test

import (
	"context"
	"database/sql"
	"testing"

	"negative-ion-respirator/backend/internal/model"
	"negative-ion-respirator/backend/internal/service"
)

// mockDeviceRepo implements service.DeviceRepo for testing.
type mockDeviceRepo struct {
	devices map[int64]*model.Device
	configs map[int64]*model.DeviceConfig
}

func (m *mockDeviceRepo) FindByID(_ context.Context, id int64) (*model.Device, error) {
	if d, ok := m.devices[id]; ok {
		return d, nil
	}
	return nil, sql.ErrNoRows
}

func (m *mockDeviceRepo) FindBySN(_ context.Context, sn string) (*model.Device, error) {
	for _, d := range m.devices {
		if d.DeviceSN == sn {
			return d, nil
		}
	}
	return nil, sql.ErrNoRows
}

func (m *mockDeviceRepo) List(_ context.Context, offset, limit int) ([]model.Device, int, error) {
	var result []model.Device
	for _, d := range m.devices {
		result = append(result, *d)
	}
	// Apply naive offset/limit
	start := offset
	if start > len(result) {
		return nil, len(result), nil
	}
	end := start + limit
	if end > len(result) {
		end = len(result)
	}
	return result[start:end], len(result), nil
}

func (m *mockDeviceRepo) UpdateStatus(_ context.Context, id int64, status string) error {
	if d, ok := m.devices[id]; ok {
		d.Status = status
		return nil
	}
	return sql.ErrNoRows
}

func (m *mockDeviceRepo) Create(_ context.Context, d *model.Device) error {
	d.ID = int64(len(m.devices) + 1)
	m.devices[d.ID] = d
	return nil
}

func (m *mockDeviceRepo) GetConfig(_ context.Context, deviceID int64) (*model.DeviceConfig, error) {
	if c, ok := m.configs[deviceID]; ok {
		return c, nil
	}
	return nil, sql.ErrNoRows
}

func (m *mockDeviceRepo) UpsertConfig(_ context.Context, cfg *model.DeviceConfig) error {
	m.configs[cfg.DeviceID] = cfg
	return nil
}

func (m *mockDeviceRepo) GetRegionConfig(_ context.Context, region, season string) (*model.RegionConfig, error) {
	return &model.RegionConfig{MaxHeatTemp: 80, TargetOutTemp: 35}, nil
}

func (m *mockDeviceRepo) UpsertRegionConfig(_ context.Context, c *model.RegionConfig) error {
	return nil
}

func TestDeviceService_UpdateConfig_ValidatesRange(t *testing.T) {
	repo := &mockDeviceRepo{
		devices: map[int64]*model.Device{
			1: {ID: 1, DeviceSN: "test-001", Status: "offline"},
		},
		configs: make(map[int64]*model.DeviceConfig),
	}
	svc := service.NewDeviceService(repo, nil) // nil mqtt client - won't send commands

	ctx := context.Background()

	tests := []struct {
		name      string
		maxHeat   int
		targetOut int
		wantErr   bool
	}{
		{"valid config", 80, 35, false},
		{"max_heat too high", 90, 35, true},
		{"max_heat negative", -1, 35, true},
		{"target_out too low", 80, 20, true},
		{"target_out too high", 80, 50, true},
		{"boundary max_heat 0", 0, 35, false},
		{"boundary max_heat 80", 80, 35, false},
		{"boundary target_out 30", 80, 30, false},
		{"boundary target_out 40", 80, 40, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := svc.UpdateConfig(ctx, 1, tt.maxHeat, tt.targetOut)
			if (err != nil) != tt.wantErr {
				t.Errorf("UpdateConfig(%d, %d) error=%v, wantErr=%v", tt.maxHeat, tt.targetOut, err, tt.wantErr)
			}
		})
	}
}

func TestDeviceService_Start_DeviceNotFound(t *testing.T) {
	repo := &mockDeviceRepo{
		devices: make(map[int64]*model.Device),
		configs: make(map[int64]*model.DeviceConfig),
	}
	svc := service.NewDeviceService(repo, nil)

	err := svc.Start(context.Background(), 999, "test-tid")
	if err == nil {
		t.Error("expected error for non-existent device")
	}
}

func TestDeviceService_Register(t *testing.T) {
	repo := &mockDeviceRepo{
		devices: make(map[int64]*model.Device),
		configs: make(map[int64]*model.DeviceConfig),
	}
	svc := service.NewDeviceService(repo, nil)

	d, err := svc.Register(context.Background(), "SN-001", "Test Device", "default")
	if err != nil {
		t.Fatalf("Register failed: %v", err)
	}
	if d.DeviceSN != "SN-001" {
		t.Errorf("expected SN 'SN-001', got '%s'", d.DeviceSN)
	}
	if d.Status != "offline" {
		t.Errorf("expected status 'offline', got '%s'", d.Status)
	}
}

func TestDeviceService_ListDevices(t *testing.T) {
	repo := &mockDeviceRepo{
		devices: map[int64]*model.Device{
			1: {ID: 1, DeviceSN: "sn-001", DeviceName: "Device 1", Status: "online"},
			2: {ID: 2, DeviceSN: "sn-002", DeviceName: "Device 2", Status: "offline"},
		},
		configs: make(map[int64]*model.DeviceConfig),
	}
	svc := service.NewDeviceService(repo, nil)

	devices, total, err := svc.ListDevices(context.Background(), 1, 10)
	if err != nil {
		t.Fatalf("ListDevices failed: %v", err)
	}
	if total != 2 {
		t.Errorf("expected total 2, got %d", total)
	}
	if len(devices) != 2 {
		t.Errorf("expected 2 devices, got %d", len(devices))
	}
}

func TestDeviceService_GetDevice_Found(t *testing.T) {
	repo := &mockDeviceRepo{
		devices: map[int64]*model.Device{
			1: {ID: 1, DeviceSN: "sn-001", DeviceName: "Test Device", Status: "online"},
		},
		configs: make(map[int64]*model.DeviceConfig),
	}
	svc := service.NewDeviceService(repo, nil)

	d, err := svc.GetDevice(context.Background(), 1)
	if err != nil {
		t.Fatalf("GetDevice failed: %v", err)
	}
	if d.DeviceSN != "sn-001" {
		t.Errorf("expected sn 'sn-001', got '%s'", d.DeviceSN)
	}
}

func TestDeviceService_GetDevice_NotFound(t *testing.T) {
	repo := &mockDeviceRepo{
		devices: make(map[int64]*model.Device),
		configs: make(map[int64]*model.DeviceConfig),
	}
	svc := service.NewDeviceService(repo, nil)

	_, err := svc.GetDevice(context.Background(), 999)
	if err == nil {
		t.Error("expected error for non-existent device")
	}
}
