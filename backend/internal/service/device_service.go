package service

import (
	"context"
	"fmt"
	"negative-ion-respirator/backend/internal/model"
	"negative-ion-respirator/backend/internal/mqtt"
	"negative-ion-respirator/backend/internal/repository"
)

type DeviceService struct {
	repo *repository.DeviceRepo
	mqtt *mqtt.Client
}

func NewDeviceService(repo *repository.DeviceRepo, mqttClient *mqtt.Client) *DeviceService {
	return &DeviceService{repo: repo, mqtt: mqttClient}
}

func (s *DeviceService) GetDevice(ctx context.Context, id int64) (*model.Device, error) {
	return s.repo.FindByID(ctx, id)
}

func (s *DeviceService) ListDevices(ctx context.Context, page, pageSize int) ([]model.Device, int, error) {
	offset := (page - 1) * pageSize
	return s.repo.List(ctx, offset, pageSize)
}

func (s *DeviceService) Start(ctx context.Context, deviceID int64, tid string) error {
	d, err := s.repo.FindByID(ctx, deviceID)
	if err != nil {
		return fmt.Errorf("device not found: %w", err)
	}
	if d.Status == "running" {
		return fmt.Errorf("device %d is already running", deviceID)
	}

	cfg, err := s.repo.GetConfig(ctx, deviceID)
	maxHeat, targetOut := 80, 35
	if err == nil {
		maxHeat = cfg.MaxHeatTemp
		targetOut = cfg.TargetOutTemp
	}

	cmd := model.DeviceCmd{
		Cmd: "start", TID: tid, MaxHeat: maxHeat, TargetOut: targetOut,
	}
	return s.mqtt.SendCmd(ctx, deviceID, cmd)
}

func (s *DeviceService) Stop(ctx context.Context, deviceID int64, tid string) error {
	cmd := model.DeviceCmd{Cmd: "stop", TID: tid}
	return s.mqtt.SendCmd(ctx, deviceID, cmd)
}

func (s *DeviceService) UpdateConfig(ctx context.Context, deviceID int64, maxHeat, targetOut int) error {
	if maxHeat < 0 || maxHeat > 80 {
		return fmt.Errorf("max_heat must be 0-80, got %d", maxHeat)
	}
	if targetOut < 30 || targetOut > 40 {
		return fmt.Errorf("target_out must be 30-40, got %d", targetOut)
	}

	cfg := &model.DeviceConfig{
		DeviceID:      deviceID,
		MaxHeatTemp:   maxHeat,
		TargetOutTemp: targetOut,
	}
	if err := s.repo.UpsertConfig(ctx, cfg); err != nil {
		return fmt.Errorf("save config: %w", err)
	}

	d, err := s.repo.FindByID(ctx, deviceID)
	if err != nil {
		return err
	}
	if d.Status == "running" {
		cmd := model.DeviceCmd{Cmd: "config", TID: "", MaxHeat: maxHeat, TargetOut: targetOut}
		return s.mqtt.SendCmd(ctx, deviceID, cmd)
	}
	return nil
}

func (s *DeviceService) Register(ctx context.Context, sn, name, region string) (*model.Device, error) {
	d := &model.Device{
		DeviceSN:   sn,
		DeviceName: name,
		RegionCode: region,
		MqttTopic:  fmt.Sprintf("device/%s", sn),
		Status:     "offline",
	}
	if err := s.repo.Create(ctx, d); err != nil {
		return nil, err
	}
	return d, nil
}

func (s *DeviceService) ListRegions(ctx context.Context) ([]string, error) {
	return []string{"default", "north", "south"}, nil
}
