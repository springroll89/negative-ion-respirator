#include "mqtt_client.h"
#include "mqtt_client.h"
#include "esp_log.h"
#include "esp_event.h"
#include "cJSON.h"
#include <string.h>

static const char *TAG = "mqtt";
static esp_mqtt_client_handle_t client = NULL;
static char device_sn[64] = {0};
static char cmd_topic[128] = {0};
static char status_topic[128] = {0};
static char heartbeat_topic[128] = {0};
static char event_topic[128] = {0};
static mqtt_cmd_callback_t cmd_cb = NULL;

static void mqtt_event_handler(void *arg, esp_event_base_t base,
                                int32_t event_id, void *event_data) {
    esp_mqtt_event_handle_t evt = event_data;

    switch (event_id) {
    case MQTT_EVENT_CONNECTED:
        ESP_LOGI(TAG, "MQTT connected");
        esp_mqtt_client_subscribe(client, cmd_topic, 1);
        break;
    case MQTT_EVENT_DATA:
        if (evt->topic_len >= strlen(cmd_topic) &&
            strncmp(evt->topic, cmd_topic, strlen(cmd_topic)) == 0) {
            cJSON *root = cJSON_ParseWithLength(evt->data, evt->data_len);
            if (root) {
                cJSON *cmd = cJSON_GetObjectItem(root, "cmd");
                cJSON *tid = cJSON_GetObjectItem(root, "tid");
                cJSON *max_heat = cJSON_GetObjectItem(root, "max_heat");
                cJSON *target_out = cJSON_GetObjectItem(root, "target_out");

                if (cmd && cmd_cb) {
                    cmd_cb(
                        cmd->valuestring,
                        tid ? tid->valuestring : "",
                        max_heat ? max_heat->valueint : 80,
                        target_out ? target_out->valueint : 35
                    );
                }
                cJSON_Delete(root);
            }
        }
        break;
    case MQTT_EVENT_DISCONNECTED:
        ESP_LOGW(TAG, "MQTT disconnected");
        break;
    default:
        break;
    }
}

esp_err_t mqtt_client_init(const char *broker_url, const char *sn) {
    strncpy(device_sn, sn, sizeof(device_sn) - 1);
    snprintf(cmd_topic, sizeof(cmd_topic), "device/%s/cmd", sn);
    snprintf(status_topic, sizeof(status_topic), "device/%s/status", sn);
    snprintf(heartbeat_topic, sizeof(heartbeat_topic), "device/%s/heartbeat", sn);
    snprintf(event_topic, sizeof(event_topic), "device/%s/event", sn);

    esp_mqtt_client_config_t cfg = {
        .broker.address.uri = broker_url,
        .credentials.client_id = sn,
        .session.keepalive = 60,
        .session.last_will = {
            .topic = heartbeat_topic,
            .msg = "{\"status\":\"offline\"}",
            .qos = 1,
            .retain = 1,
        },
    };

    client = esp_mqtt_client_init(&cfg);
    esp_mqtt_client_register_event(client, ESP_EVENT_ANY_ID, mqtt_event_handler, NULL);
    return ESP_OK;
}

esp_err_t mqtt_client_start(void) {
    return esp_mqtt_client_start(client);
}

esp_err_t mqtt_client_stop(void) {
    return esp_mqtt_client_stop(client);
}

esp_err_t mqtt_publish_status(const char *status, float heat_temp, float out_temp,
                               bool ion_ok, uint32_t uptime) {
    cJSON *root = cJSON_CreateObject();
    cJSON_AddStringToObject(root, "status", status);
    cJSON_AddNumberToObject(root, "heat_temp", heat_temp);
    cJSON_AddNumberToObject(root, "out_temp", out_temp);
    cJSON_AddBoolToObject(root, "ion_ok", ion_ok);
    cJSON_AddNumberToObject(root, "uptime", uptime);

    char *str = cJSON_PrintUnformatted(root);
    int msg_id = esp_mqtt_client_publish(client, status_topic, str, 0, 1, 0);
    free(str);
    cJSON_Delete(root);
    return msg_id >= 0 ? ESP_OK : ESP_FAIL;
}

esp_err_t mqtt_publish_heartbeat(int rssi, const char *conn_type, const char *version) {
    cJSON *root = cJSON_CreateObject();
    cJSON_AddNumberToObject(root, "rssi", rssi);
    cJSON_AddStringToObject(root, "conn_type", conn_type);
    cJSON_AddStringToObject(root, "version", version);
    cJSON_AddNumberToObject(root, "heap", esp_get_free_heap_size());

    char *str = cJSON_PrintUnformatted(root);
    int msg_id = esp_mqtt_client_publish(client, heartbeat_topic, str, 0, 0, 0);
    free(str);
    cJSON_Delete(root);
    return msg_id >= 0 ? ESP_OK : ESP_FAIL;
}

esp_err_t mqtt_publish_event(const char *event, float value, float limit, const char *action) {
    cJSON *root = cJSON_CreateObject();
    cJSON_AddStringToObject(root, "event", event);
    cJSON_AddNumberToObject(root, "value", value);
    cJSON_AddNumberToObject(root, "limit", limit);
    cJSON_AddStringToObject(root, "action", action);

    char *str = cJSON_PrintUnformatted(root);
    int msg_id = esp_mqtt_client_publish(client, event_topic, str, 0, 1, 0);
    free(str);
    cJSON_Delete(root);
    return msg_id >= 0 ? ESP_OK : ESP_FAIL;
}

void mqtt_set_cmd_callback(mqtt_cmd_callback_t cb) {
    cmd_cb = cb;
}
