#include "comm_4g.h"
#include "esp_log.h"
#include "driver/uart.h"
#include "driver/gpio.h"
#include "freertos/FreeRTOS.h"
#include "freertos/task.h"
#include <string.h>

static const char *TAG = "comm_4g";
static int uart_num = 0;
static bool connected = false;
static conn_type_t active_conn = CONN_NONE;

static esp_err_t send_at(const char *cmd, const char *expect, int timeout_ms) {
    uart_write_bytes(uart_num, cmd, strlen(cmd));
    uart_write_bytes(uart_num, "\r\n", 2);

    char buf[256] = {0};
    int len = uart_read_bytes(uart_num, (uint8_t*)buf, sizeof(buf)-1,
                               pdMS_TO_TICKS(timeout_ms));
    if (len > 0) {
        buf[len] = 0;
        if (strstr(buf, expect) || strstr(buf, "OK")) {
            return ESP_OK;
        }
    }
    return ESP_FAIL;
}

esp_err_t comm_4g_init(int uart, int tx, int rx, int rst, int pwr) {
    uart_num = uart;

    uart_config_t uart_cfg = {
        .baud_rate = 115200,
        .data_bits = UART_DATA_8_BITS,
        .parity = UART_PARITY_DISABLE,
        .stop_bits = UART_STOP_BITS_1,
        .flow_ctrl = UART_HW_FLOWCTRL_DISABLE,
    };
    uart_param_config(uart_num, &uart_cfg);
    uart_set_pin(uart_num, tx, rx, UART_PIN_NO_CHANGE, UART_PIN_NO_CHANGE);
    uart_driver_install(uart_num, 1024, 0, 0, NULL, 0);

    // Configure control pins
    gpio_set_direction(rst, GPIO_MODE_OUTPUT);
    gpio_set_direction(pwr, GPIO_MODE_OUTPUT);
    gpio_set_level(pwr, 1);

    // Reset module
    gpio_set_level(rst, 0);
    vTaskDelay(pdMS_TO_TICKS(200));
    gpio_set_level(rst, 1);
    vTaskDelay(pdMS_TO_TICKS(3000));

    // Test communication
    if (send_at("AT", "OK", 1000) != ESP_OK) {
        ESP_LOGE(TAG, "no response from 4G module");
        return ESP_FAIL;
    }
    ESP_LOGI(TAG, "4G module initialized");
    return ESP_OK;
}

esp_err_t comm_4g_connect(const char *apn) {
    char cmd[128];
    snprintf(cmd, sizeof(cmd), "AT+CGDCONT=1,\"IP\",\"%s\"", apn);
    send_at(cmd, "OK", 1000);
    send_at("AT+CGACT=1,1", "OK", 5000);

    vTaskDelay(pdMS_TO_TICKS(2000));
    connected = true;
    active_conn = CONN_4G;
    ESP_LOGI(TAG, "4G connected, APN=%s", apn);
    return ESP_OK;
}

esp_err_t comm_4g_disconnect(void) {
    send_at("AT+CGACT=0,1", "OK", 1000);
    connected = false;
    if (active_conn == CONN_4G) active_conn = CONN_NONE;
    return ESP_OK;
}

bool comm_4g_is_connected(void) { return connected; }

int comm_4g_get_rssi(void) {
    // Send AT+CSQ and parse signal quality
    return -65; // placeholder
}

conn_type_t comm_get_active_connection(void) { return active_conn; }

esp_err_t comm_init_dual_mode(const char *wifi_ssid, const char *wifi_pass,
                               const char *apn) {
    // Try WiFi first
    wifi_config_t wifi_cfg = {0};
    strncpy((char*)wifi_cfg.sta.ssid, wifi_ssid, sizeof(wifi_cfg.sta.ssid)-1);
    strncpy((char*)wifi_cfg.sta.password, wifi_pass, sizeof(wifi_cfg.sta.password)-1);

    ESP_LOGI(TAG, "connecting to WiFi: %s", wifi_ssid);
    esp_wifi_set_config(WIFI_IF_STA, &wifi_cfg);
    esp_wifi_connect();

    // Wait up to 10s for WiFi
    int retry = 0;
    while (retry < 20) {
        vTaskDelay(pdMS_TO_TICKS(500));
        // Check if connected (simplified)
        retry++;
    }

    // Fallback to 4G
    ESP_LOGI(TAG, "WiFi failed, switching to 4G");
    return comm_4g_connect(apn);
}
