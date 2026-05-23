#include "led_indicator.h"
#include "driver/gpio.h"
#include "freertos/FreeRTOS.h"
#include "freertos/timers.h"

static int gpio_power = 0, gpio_work = 0;
static TimerHandle_t blink_timer = NULL;
static int blink_state = 0;

static void blink_callback(TimerHandle_t timer) {
    blink_state = !blink_state;
    gpio_set_level(gpio_work, blink_state);
}

esp_err_t led_indicator_init(int power_pin, int work_pin) {
    gpio_power = power_pin;
    gpio_work = work_pin;

    gpio_config_t io_conf = {
        .mode = GPIO_MODE_OUTPUT,
        .pull_up_en = GPIO_PULLUP_DISABLE,
        .pull_down_en = GPIO_PULLDOWN_DISABLE,
        .intr_type = GPIO_INTR_DISABLE,
    };

    io_conf.pin_bit_mask = (1ULL << gpio_power);
    gpio_config(&io_conf);
    gpio_set_level(gpio_power, 1); // power LED always on

    io_conf.pin_bit_mask = (1ULL << gpio_work);
    gpio_config(&io_conf);
    gpio_set_level(gpio_work, 0); // work LED off

    blink_timer = xTimerCreate("blink", pdMS_TO_TICKS(1000), pdTRUE, NULL, blink_callback);
    return ESP_OK;
}

esp_err_t led_indicator_update(device_state_t state) {
    if (blink_timer) xTimerStop(blink_timer, 0);

    switch (state) {
    case STATE_IDLE:
    case STATE_DONE:
        gpio_set_level(gpio_work, 0); // off
        break;
    case STATE_HEATING:
        // 1Hz blink
        if (blink_timer) {
            xTimerChangePeriod(blink_timer, pdMS_TO_TICKS(500), 0);
            xTimerStart(blink_timer, 0);
        }
        break;
    case STATE_RUNNING:
        gpio_set_level(gpio_work, 1); // solid on
        break;
    case STATE_ERROR:
        // 5Hz fast blink
        if (blink_timer) {
            xTimerChangePeriod(blink_timer, pdMS_TO_TICKS(100), 0);
            xTimerStart(blink_timer, 0);
        }
        break;
    }
    return ESP_OK;
}
