package service

import (
	"context"
	"fmt"
	"negative-ion-respirator/backend/internal/model"
	"negative-ion-respirator/backend/internal/repository"
	"time"
)

type OrderService struct {
	repo      *repository.OrderRepo
	userRepo  *repository.UserRepo
	deviceSvc *DeviceService
}

func NewOrderService(repo *repository.OrderRepo, userRepo *repository.UserRepo, deviceSvc *DeviceService) *OrderService {
	return &OrderService{repo: repo, userRepo: userRepo, deviceSvc: deviceSvc}
}

func (s *OrderService) Create(ctx context.Context, req model.CreateOrderReq) (*model.Order, error) {
	_, err := s.deviceSvc.GetDevice(ctx, req.DeviceID)
	if err != nil {
		return nil, fmt.Errorf("device %d not found", req.DeviceID)
	}

	user, err := s.userRepo.FindByID(ctx, req.UserID)
	if err != nil || user == nil {
		if user == nil {
			user = &model.User{OpenID: req.OpenID, Balance: 0}
			if err := s.userRepo.Create(ctx, user); err != nil {
				return nil, fmt.Errorf("create user: %w", err)
			}
		} else {
			return nil, fmt.Errorf("user lookup failed: %w", err)
		}
	}

	o := &model.Order{UserID: user.ID, DeviceID: req.DeviceID}
	if err := s.repo.Create(ctx, o); err != nil {
		return nil, fmt.Errorf("create order: %w", err)
	}
	return o, nil
}

func (s *OrderService) Start(ctx context.Context, tid string) error {
	o, err := s.repo.FindByTID(ctx, tid)
	if err != nil {
		return fmt.Errorf("order not found: %w", err)
	}
	if o.Status != "pending" {
		return fmt.Errorf("order %s is not pending (status=%s)", tid, o.Status)
	}

	if err := s.deviceSvc.Start(ctx, o.DeviceID, tid); err != nil {
		return fmt.Errorf("start device: %w", err)
	}

	return s.repo.MarkStarted(ctx, tid)
}

func (s *OrderService) Stop(ctx context.Context, tid string) (*model.Order, error) {
	o, err := s.repo.FindByTID(ctx, tid)
	if err != nil {
		return nil, fmt.Errorf("order not found: %w", err)
	}
	if o.Status != "active" {
		return nil, fmt.Errorf("order %s is not active", tid)
	}

	if err := s.deviceSvc.Stop(ctx, o.DeviceID, tid); err != nil {
		return nil, fmt.Errorf("stop device: %w", err)
	}

	durationSec := int(time.Since(*o.StartTime).Seconds())
	amount := int64(durationSec) * 1

	if err := s.repo.Settle(ctx, tid, durationSec, amount); err != nil {
		return nil, fmt.Errorf("settle order: %w", err)
	}

	o.Duration = durationSec
	o.Amount = amount
	o.Status = "completed"
	return o, nil
}

func (s *OrderService) FindByTID(ctx context.Context, tid string) (*model.Order, error) {
	return s.repo.FindByTID(ctx, tid)
}
