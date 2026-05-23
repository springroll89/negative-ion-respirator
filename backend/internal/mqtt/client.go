package mqtt

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"sync"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"negative-ion-respirator/backend/internal/model"
	"negative-ion-respirator/backend/internal/repository"
)

type Client struct {
	client        mqtt.Client
	deviceRepo    *repository.DeviceRepo
	telemetryRepo *repository.TelemetryRepo
	orderRepo     *repository.OrderRepo
	pending       map[string]chan []byte
	mu            sync.RWMutex
}

func NewClient(brokerURL, clientID string, deviceRepo *repository.DeviceRepo, telemetryRepo *repository.TelemetryRepo, orderRepo *repository.OrderRepo) (*Client, error) {
	c := &Client{
		deviceRepo:    deviceRepo,
		telemetryRepo: telemetryRepo,
		orderRepo:     orderRepo,
		pending:       make(map[string]chan []byte),
	}

	opts := mqtt.NewClientOptions().
		AddBroker(brokerURL).
		SetClientID(clientID).
		SetAutoReconnect(true).
		SetMaxReconnectInterval(10).
		SetOnConnectHandler(c.onConnect)

	c.client = mqtt.NewClient(opts)
	if token := c.client.Connect(); token.Wait() && token.Error() != nil {
		return nil, fmt.Errorf("mqtt connect: %w", token.Error())
	}

	return c, nil
}

func (c *Client) onConnect(client mqtt.Client) {
	log.Println("MQTT connected, subscribing to device topics...")
	client.Subscribe("device/+/status", 1, c.handleStatus)
	client.Subscribe("device/+/heartbeat", 0, c.handleHeartbeat)
	client.Subscribe("device/+/event", 1, c.handleEvent)
}

func (c *Client) SendCmd(ctx context.Context, deviceID int64, cmd model.DeviceCmd) error {
	d, err := c.deviceRepo.FindByID(ctx, deviceID)
	if err != nil {
		return fmt.Errorf("find device: %w", err)
	}

	payload, _ := json.Marshal(cmd)
	topic := fmt.Sprintf("device/%s/cmd", d.DeviceSN)

	token := c.client.Publish(topic, 1, false, payload)
	if token.Wait(); token.Error() != nil {
		return fmt.Errorf("publish: %w", token.Error())
	}
	return nil
}

func (c *Client) handleStatus(client mqtt.Client, msg mqtt.Message) {
	var status model.DeviceStatus
	if err := json.Unmarshal(msg.Payload(), &status); err != nil {
		log.Printf("bad status message: %v", err)
		return
	}
	log.Printf("device status: status=%s heat=%.1f out=%.1f ion=%v",
		status.Status, status.HeatTemp, status.OutTemp, status.IonOK)
}

func (c *Client) handleHeartbeat(client mqtt.Client, msg mqtt.Message) {
	var hb model.DeviceHeartbeat
	if err := json.Unmarshal(msg.Payload(), &hb); err != nil {
		log.Printf("bad heartbeat message: %v", err)
		return
	}
	log.Printf("device heartbeat: rssi=%d heap=%d conn=%s ver=%s",
		hb.RSSI, hb.Heap, hb.ConnType, hb.Version)
}

func (c *Client) handleEvent(client mqtt.Client, msg mqtt.Message) {
	var evt model.DeviceEvent
	if err := json.Unmarshal(msg.Payload(), &evt); err != nil {
		log.Printf("bad event message: %v", err)
		return
	}
	log.Printf("ALERT device event: event=%s value=%.1f action=%s",
		evt.Event, evt.Value, evt.Action)
}

func (c *Client) Close() {
	c.client.Disconnect(250)
}
