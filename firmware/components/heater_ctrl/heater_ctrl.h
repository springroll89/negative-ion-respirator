#ifndef HEATER_CTRL_H
#define HEATER_CTRL_H

#include "esp_err.h"
#include <stdint.h>
#include <stdbool.h>

esp_err_t heater_ctrl_init(int gpio_pwm, int gpio_adc_heat, int gpio_adc_out);
esp_err_t heater_ctrl_set_target(int heat_max, int out_target);
esp_err_t heater_ctrl_get_temps(float *heat_temp, float *out_temp);
esp_err_t heater_ctrl_run(void);    // call periodically (e.g. every 500ms)
esp_err_t heater_ctrl_stop(void);
esp_err_t heater_ctrl_emergency_shutdown(void);

#endif
