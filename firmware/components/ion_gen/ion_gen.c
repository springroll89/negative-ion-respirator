#include "ion_gen.h"
#include "esp_log.h"
#include "driver/gpio.h"

static const char *TAG = "ion_gen";
static int gpio_enable = 0, gpio_fault = 0;
static bool running = false;

esp_err_t ion_gen_init(int enable_pin, int fault_pin) {
    gpio_enable = enable_pin;
    gpio_fault = fault_pin;

    gpio_config_t io_conf = {
        .pin_bit_mask = (1ULL << gpio_enable),
        .mode = GPIO_MODE_OUTPUT,
        .pull_up_en = GPIO_PULLUP_DISABLE,
        .pull_down_en = GPIO_PULLDOWN_DISABLE,
        .intr_type = GPIO_INTR_DISABLE,
    };
    gpio_config(&io_conf);

    io_conf.pin_bit_mask = (1ULL << gpio_fault);
    io_conf.mode = GPIO_MODE_INPUT;
    io_conf.pull_up_en = GPIO_PULLUP_ENABLE;
    gpio_config(&io_conf);

    gpio_set_level(gpio_enable, 0);
    ESP_LOGI(TAG, "ion generator initialized: enable=%d fault=%d", enable_pin, fault_pin);
    return ESP_OK;
}

esp_err_t ion_gen_start(void) {
    gpio_set_level(gpio_enable, 1);
    running = true;
    ESP_LOGI(TAG, "ion generator started");
    return ESP_OK;
}

esp_err_t ion_gen_stop(void) {
    gpio_set_level(gpio_enable, 0);
    running = false;
    ESP_LOGI(TAG, "ion generator stopped");
    return ESP_OK;
}

bool ion_gen_is_ok(void) {
    if (!running) return true;
    return gpio_get_level(gpio_fault) == 1; // pulled high = OK
}
