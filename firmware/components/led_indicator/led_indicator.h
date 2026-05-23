#ifndef LED_INDICATOR_H
#define LED_INDICATOR_H

#include "esp_err.h"
#include "../state_machine/state_machine.h"

esp_err_t led_indicator_init(int gpio_power, int gpio_work);
esp_err_t led_indicator_update(device_state_t state);

#endif
