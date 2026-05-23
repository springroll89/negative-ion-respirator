package model

import "time"

type BatchTask struct {
	ID         int64      `json:"id"`
	TaskType   string     `json:"task_type"`
	TargetType string     `json:"target_type"`
	TargetIDs  []int64    `json:"target_ids"`
	ConfigJSON []byte     `json:"config_json"`
	Status     string     `json:"status"`
	Progress   int        `json:"progress"`
	Total      int        `json:"total"`
	CreatedAt  time.Time  `json:"created_at"`
	FinishedAt *time.Time `json:"finished_at"`
}
