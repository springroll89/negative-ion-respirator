#ifndef ION_GEN_H
#define ION_GEN_H

#include "esp_err.h"
#include <stdbool.h>

esp_err_t ion_gen_init(int gpio_enable, int gpio_fault_detect);
esp_err_t ion_gen_start(void);
esp_err_t ion_gen_stop(void);
bool ion_gen_is_ok(void);

#endif
