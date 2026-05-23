package model

import "time"

type DeviceLog struct {
	Time      time.Time `json:"time"`
	DeviceID  int64     `json:"device_id"`
	Status    string    `json:"status"`
	HeatTemp  float64   `json:"heat_temp"`
	OutTemp   float64   `json:"out_temp"`
	IonOK     bool      `json:"ion_ok"`
	EventType string    `json:"event_type,omitempty"`
	EventData []byte    `json:"event_data,omitempty"`
}
