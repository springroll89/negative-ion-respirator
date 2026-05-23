#include <stdio.h>
#include "freertos/FreeRTOS.h"
#include "freertos/task.h"
#include "esp_system.h"
#include "esp_log.h"
#include "esp_wifi.h"
#include "nvs_flash.h"

#include "state_machine.h"
#include "heater_ctrl.h"
#include "ion_gen.h"
#include "led_indicator.h"
#include "safety_watchdog.h"
#include "mqtt_client.h"

static const char *TAG = "main";
static char current_tid[64] = {0};
static uint32_t uptime_seconds = 0;
static bool device_running = false;

static void on_state_change(device_state_t old, device_state_t new_state) {
    led_indicator_update(new_state);
}

static void on_cmd_received(const char *cmd, const char *tid,
                             int max_heat, int target_out) {
    ESP_LOGI(TAG, "cmd received: %s, tid=%s, max_heat=%d, target_out=%d",
             cmd, tid, max_heat, target_out);

    if (strcmp(cmd, "start") == 0) {
        if (state_machine_get_state() != STATE_IDLE) {
            ESP_LOGW(TAG, "cannot start: not in idle state");
            return;
        }
        strncpy(current_tid, tid, sizeof(current_tid) - 1);
        heater_ctrl_set_target(max_heat, target_out);
        state_machine_transition(STATE_HEATING);
        ion_gen_start();
        heater_ctrl_run(); // will be called periodically by timer
        device_running = true;
    } else if (strcmp(cmd, "stop") == 0) {
        heater_ctrl_stop();
        ion_gen_stop();
        state_machine_transition(STATE_DONE);
        device_running = false;
        state_machine_transition(STATE_IDLE);
    } else if (strcmp(cmd, "config") == 0) {
        heater_ctrl_set_target(max_heat, target_out);
    }
}

void app_main(void) {
    esp_err_t ret = nvs_flash_init();
    if (ret == ESP_ERR_NVS_NO_FREE_PAGES || ret == ESP_ERR_NVS_NEW_VERSION_FOUND) {
        nvs_flash_erase();
        nvs_flash_init();
    }

    ESP_LOGI(TAG, "Negative Ion Respirator Firmware v1.0");

    // Initialize components
    state_machine_init();
    state_machine_set_callback(on_state_change);

    // GPIO config: PWM=GPIO25, ADC_HEAT=GPIO36, ADC_OUT=GPIO39, ION_EN=GPIO26, ION_FAULT=GPIO27
    // LED_POWER=GPIO32, LED_WORK=GPIO33, HW_SHUTDOWN=GPIO14
    heater_ctrl_init(25, 36, 39);
    ion_gen_init(26, 27);
    led_indicator_init(32, 33);
    safety_watchdog_init();
    safety_watchdog_enable_hw_protection(14);

    // Init Wi-Fi
    // ... (Wi-Fi init code depends on project config)

    // Init MQTT
    mqtt_client_init("mqtt://192.168.1.100:1883", "ion-respirator-001");
    mqtt_set_cmd_callback(on_cmd_received);
    mqtt_client_start();

    // Main loop
    TickType_t last_status = xTaskGetTickCount();
    TickType_t last_heartbeat = xTaskGetTickCount();
    TickType_t last_run = xTaskGetTickCount();

    while (1) {
        TickType_t now = xTaskGetTickCount();

        // Feed watchdog
        safety_watchdog_feed();

        // Run heater control every 500ms
        if (now - last_run >= pdMS_TO_TICKS(500)) {
            if (device_running) {
                heater_ctrl_run();
                // Check if heating is done
                if (state_machine_get_state() == STATE_HEATING) {
                    float heat, out;
                    heater_ctrl_get_temps(&heat, &out);
                    if (out >= 30.0) { // target output temp reached
                        state_machine_transition(STATE_RUNNING);
                    }
                }
            }
            last_run = now;
        }

        // Publish status every 5s
        if (now - last_status >= pdMS_TO_TICKS(5000)) {
            float heat_temp, out_temp;
            heater_ctrl_get_temps(&heat_temp, &out_temp);
            mqtt_publish_status(
                state_machine_get_state_name(),
                heat_temp, out_temp,
                ion_gen_is_ok(),
                ++uptime_seconds / 5
            );
            last_status = now;
        }

        // Publish heartbeat every 30s
        if (now - last_heartbeat >= pdMS_TO_TICKS(30000)) {
            mqtt_publish_heartbeat(0, "wifi", "1.0.0");
            last_heartbeat = now;
        }

        // Check for faults
        if (device_running && !ion_gen_is_ok()) {
            ESP_LOGE(TAG, "ion generator fault detected!");
            mqtt_publish_event("ion_fail", 0, 0, "auto_shutdown");
            heater_ctrl_stop();
            ion_gen_stop();
            state_machine_transition(STATE_ERROR);
            device_running = false;
        }

        vTaskDelay(pdMS_TO_TICKS(100));
    }
}
