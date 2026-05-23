#ifndef SAFETY_WATCHDOG_H
#define SAFETY_WATCHDOG_H

#include "esp_err.h"

esp_err_t safety_watchdog_init(void);
esp_err_t safety_watchdog_feed(void);
esp_err_t safety_watchdog_enable_hw_protection(int gpio_over_temp_shutdown);

#endif
