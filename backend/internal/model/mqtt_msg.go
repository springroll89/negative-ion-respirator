package model

type DeviceCmd struct {
	Cmd       string `json:"cmd"`
	TID       string `json:"tid"`
	MaxHeat   int    `json:"max_heat,omitempty"`
	TargetOut int    `json:"target_out,omitempty"`
}

type DeviceStatus struct {
	Status   string  `json:"status"`
	TID      string  `json:"tid,omitempty"`
	HeatTemp float64 `json:"heat_temp"`
	OutTemp  float64 `json:"out_temp"`
	IonOK    bool    `json:"ion_ok"`
	Uptime   int     `json:"uptime"`
}

type DeviceHeartbeat struct {
	RSSI     int    `json:"rssi"`
	Heap     uint32 `json:"heap"`
	ConnType string `json:"conn_type"`
	Version  string `json:"version"`
}

type DeviceEvent struct {
	Event  string  `json:"event"`
	Value  float64 `json:"value,omitempty"`
	Limit  float64 `json:"limit,omitempty"`
	Action string  `json:"action"`
}
