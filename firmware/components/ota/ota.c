#include "ota.h"
#include "esp_log.h"
#include "esp_ota_ops.h"
#include "esp_https_ota.h"
#include "nvs_flash.h"
#include "nvs.h"

static const char *TAG = "ota";
static char current_version[32] = "1.0.0";
static const esp_partition_t *update_partition = NULL;

esp_err_t ota_init(const char *version) {
    strncpy(current_version, version, sizeof(current_version) - 1);

    update_partition = esp_ota_get_next_update_partition(NULL);
    if (update_partition == NULL) {
        ESP_LOGE(TAG, "no OTA partition found");
        return ESP_FAIL;
    }
    ESP_LOGI(TAG, "OTA initialized, current version=%s, update partition=%s",
             current_version, update_partition->label);
    return ESP_OK;
}

esp_err_t ota_check_and_update(const char *firmware_url, const char *expected_version,
                                ota_progress_callback_t progress_cb) {
    if (strcmp(expected_version, current_version) == 0) {
        ESP_LOGI(TAG, "already on version %s, skipping OTA", current_version);
        return ESP_OK;
    }

    ESP_LOGI(TAG, "starting OTA from %s to %s", current_version, expected_version);
    ESP_LOGI(TAG, "firmware URL: %s", firmware_url);

    esp_http_client_config_t http_cfg = {
        .url = firmware_url,
        .timeout_ms = 60000,
        .keep_alive_enable = true,
    };

    esp_https_ota_config_t ota_cfg = {
        .http_config = &http_cfg,
    };

    esp_err_t ret = esp_https_ota(&ota_cfg);
    if (ret == ESP_OK) {
        ESP_LOGI(TAG, "OTA update successful, restarting...");
        strncpy(current_version, expected_version, sizeof(current_version) - 1);
        // Save version to NVS
        nvs_handle_t nvs;
        nvs_open("ota", NVS_READWRITE, &nvs);
        nvs_set_str(nvs, "version", current_version);
        nvs_commit(nvs);
        nvs_close(nvs);

        vTaskDelay(pdMS_TO_TICKS(1000));
        esp_restart();
    } else {
        ESP_LOGE(TAG, "OTA update failed: %s", esp_err_to_name(ret));
        return ret;
    }
    return ESP_OK;
}

esp_err_t ota_rollback(void) {
    esp_ota_img_states_t state;
    if (esp_ota_get_state_partition(update_partition, &state) == ESP_OK) {
        if (state == ESP_OTA_IMG_PENDING_VERIFY) {
            ESP_LOGW(TAG, "rolling back OTA update");
            esp_ota_mark_app_invalid_rollback_and_reboot();
        }
    }
    return ESP_OK;
}

const char *ota_get_current_version(void) {
    return current_version;
}
