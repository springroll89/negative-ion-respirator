idf_component_register(
    SRCS "mqtt_client.c"
    INCLUDE_DIRS "."
    REQUIRES mqtt cjson esp_wifi
)
