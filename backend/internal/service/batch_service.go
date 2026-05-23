package service

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"negative-ion-respirator/backend/internal/model"
	"negative-ion-respirator/backend/internal/repository"
	"time"
)

type BatchService struct {
	repo      *repository.DeviceRepo
	deviceSvc *DeviceService
	// batch tasks stored in-memory for simplicity; production would use DB
	tasks map[int64]*model.BatchTask
}

func NewBatchService(repo *repository.DeviceRepo, deviceSvc *DeviceService) *BatchService {
	return &BatchService{
		repo:      repo,
		deviceSvc: deviceSvc,
		tasks:     make(map[int64]*model.BatchTask),
	}
}

type BatchConfigReq struct {
	TaskType      string   `json:"task_type"`       // "config" or "ota"
	TargetType    string   `json:"target_type"`      // "device", "region", "season"
	TargetIDs     []int64  `json:"target_ids,omitempty"`
	TargetRegions []string `json:"target_regions,omitempty"`
	TargetSeasons []string `json:"target_seasons,omitempty"`
	MaxHeatTemp   int      `json:"max_heat_temp"`
	TargetOutTemp int      `json:"target_out_temp"`
}

func (s *BatchService) CreateBatchConfig(ctx context.Context, req BatchConfigReq) (*model.BatchTask, error) {
	// Resolve target devices
	var deviceIDs []int64

	switch req.TargetType {
	case "device":
		deviceIDs = req.TargetIDs
	case "region":
		// Get all devices in specified regions
		devices, _, err := s.repo.List(ctx, 0, 10000)
		if err != nil {
			return nil, fmt.Errorf("list devices: %w", err)
		}
		regionSet := make(map[string]bool)
		for _, r := range req.TargetRegions {
			regionSet[r] = true
		}
		for _, d := range devices {
			if regionSet[d.RegionCode] {
				deviceIDs = append(deviceIDs, d.ID)
			}
		}
	default:
		return nil, fmt.Errorf("unsupported target_type: %s", req.TargetType)
	}

	if len(deviceIDs) == 0 {
		return nil, fmt.Errorf("no devices matched")
	}

	configJSON, _ := json.Marshal(map[string]int{
		"max_heat_temp":   req.MaxHeatTemp,
		"target_out_temp": req.TargetOutTemp,
	})

	task := &model.BatchTask{
		ID:         time.Now().UnixNano(),
		TaskType:   req.TaskType,
		TargetType: req.TargetType,
		TargetIDs:  deviceIDs,
		ConfigJSON: configJSON,
		Status:     "pending",
		Total:      len(deviceIDs),
		CreatedAt:  time.Now(),
	}
	s.tasks[task.ID] = task

	// Execute in background
	go s.executeBatchConfig(task, req.MaxHeatTemp, req.TargetOutTemp)

	return task, nil
}

func (s *BatchService) executeBatchConfig(task *model.BatchTask, maxHeat, targetOut int) {
	task.Status = "running"
	log.Printf("batch config started: task=%d devices=%d", task.ID, task.Total)

	for _, deviceID := range task.TargetIDs {
		err := s.deviceSvc.UpdateConfig(context.Background(), deviceID, maxHeat, targetOut)
		if err != nil {
			log.Printf("batch config: device %d failed: %v", deviceID, err)
		} else {
			task.Progress++
		}
		time.Sleep(200 * time.Millisecond) // rate limit
	}

	task.Status = "completed"
	now := time.Now()
	task.FinishedAt = &now
	log.Printf("batch config completed: task=%d success=%d/%d", task.ID, task.Progress, task.Total)
}

func (s *BatchService) GetTask(taskID int64) (*model.BatchTask, error) {
	task, ok := s.tasks[taskID]
	if !ok {
		return nil, fmt.Errorf("task %d not found", taskID)
	}
	return task, nil
}

func (s *BatchService) ListTasks() []*model.BatchTask {
	var result []*model.BatchTask
	for _, t := range s.tasks {
		result = append(result, t)
	}
	return result
}
