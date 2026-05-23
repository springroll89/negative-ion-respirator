package model

import "time"

type Device struct {
	ID              int64      `json:"id"`
	DeviceSN        string     `json:"device_sn"`
	DeviceName      string     `json:"device_name"`
	RegionCode      string     `json:"region_code"`
	MqttTopic       string     `json:"mqtt_topic"`
	Status          string     `json:"status"`
	FirmwareVersion string     `json:"firmware_version"`
	LastOnline      *time.Time `json:"last_online"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
}

type DeviceConfig struct {
	ID            int64     `json:"id"`
	DeviceID      int64     `json:"device_id"`
	MaxHeatTemp   int       `json:"max_heat_temp"`
	TargetOutTemp int       `json:"target_out_temp"`
	UpdatedAt     time.Time `json:"updated_at"`
}

type RegionConfig struct {
	ID            int64     `json:"id"`
	RegionCode    string    `json:"region_code"`
	Season        string    `json:"season"`
	MaxHeatTemp   int       `json:"max_heat_temp"`
	TargetOutTemp int       `json:"target_out_temp"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}
