#ifndef MQTT_CLIENT_H
#define MQTT_CLIENT_H

#include "esp_err.h"
#include <stdint.h>
#include <stdbool.h>

typedef void (*mqtt_cmd_callback_t)(const char *cmd, const char *tid,
                                     int max_heat, int target_out);

esp_err_t mqtt_client_init(const char *broker_url, const char *device_sn);
esp_err_t mqtt_client_start(void);
esp_err_t mqtt_client_stop(void);

esp_err_t mqtt_publish_status(const char *status, float heat_temp, float out_temp,
                               bool ion_ok, uint32_t uptime);
esp_err_t mqtt_publish_heartbeat(int rssi, const char *conn_type, const char *version);
esp_err_t mqtt_publish_event(const char *event, float value, float limit, const char *action);

void mqtt_set_cmd_callback(mqtt_cmd_callback_t cb);

#endif
