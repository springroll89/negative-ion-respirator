#ifndef OTA_H
#define OTA_H

#include "esp_err.h"

typedef void (*ota_progress_callback_t)(int percent);

esp_err_t ota_init(const char *current_version);
esp_err_t ota_check_and_update(const char *firmware_url, const char *expected_version,
                                ota_progress_callback_t progress_cb);
esp_err_t ota_rollback(void);
const char *ota_get_current_version(void);

#endif
