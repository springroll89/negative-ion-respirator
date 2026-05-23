#ifndef COMM_4G_H
#define COMM_4G_H

#include "esp_err.h"
#include <stdbool.h>

typedef enum { CONN_WIFI, CONN_4G, CONN_NONE } conn_type_t;

esp_err_t comm_4g_init(int uart_num, int tx_pin, int rx_pin, int reset_pin, int power_pin);
esp_err_t comm_4g_connect(const char *apn);
esp_err_t comm_4g_disconnect(void);
bool comm_4g_is_connected(void);
int comm_4g_get_rssi(void);
conn_type_t comm_get_active_connection(void);

// Auto-switch: prefer WiFi, fallback to 4G
esp_err_t comm_init_dual_mode(const char *wifi_ssid, const char *wifi_pass,
                               const char *apn);

#endif
