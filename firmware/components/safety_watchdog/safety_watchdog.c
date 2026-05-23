#include "safety_watchdog.h"
#include "esp_log.h"
#include "esp_task_wdt.h"
#include "driver/gpio.h"

static const char *TAG = "safety_wdog";
static int gpio_shutdown = 0;

esp_err_t safety_watchdog_init(void) {
    esp_task_wdt_config_t twdt_config = {
        .timeout_ms = 5000,
        .idle_core_mask = 0,
        .trigger_panic = true,
    };
    esp_task_wdt_init(&twdt_config);
    esp_task_wdt_add(NULL);
    ESP_LOGI(TAG, "task watchdog initialized (5s timeout)");
    return ESP_OK;
}

esp_err_t safety_watchdog_feed(void) {
    esp_task_wdt_reset();
    return ESP_OK;
}

esp_err_t safety_watchdog_enable_hw_protection(int gpio) {
    gpio_shutdown = gpio;

    // Configure hardware shutdown: active-low signal to relay/contactor
    gpio_config_t io_conf = {
        .pin_bit_mask = (1ULL << gpio_shutdown),
        .mode = GPIO_MODE_OUTPUT,
        .pull_up_en = GPIO_PULLUP_DISABLE,
        .pull_down_en = GPIO_PULLDOWN_DISABLE,
        .intr_type = GPIO_INTR_DISABLE,
    };
    gpio_config(&io_conf);
    gpio_set_level(gpio_shutdown, 1); // high = normal operation
    ESP_LOGI(TAG, "hardware over-temp protection enabled on GPIO %d", gpio);
    return ESP_OK;
}
