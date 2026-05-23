idf_component_register(
    SRCS "heater_ctrl.c"
    INCLUDE_DIRS "."
    REQUIRES driver esp_adc
)
