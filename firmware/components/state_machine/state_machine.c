#include "state_machine.h"
#include "esp_log.h"

static const char *TAG = "state_machine";
static device_state_t current_state = STATE_IDLE;
static state_change_callback_t callback = NULL;

static const char *state_names[] = {
    [STATE_IDLE]    = "idle",
    [STATE_HEATING] = "heating",
    [STATE_RUNNING] = "running",
    [STATE_DONE]    = "done",
    [STATE_ERROR]   = "error",
};

void state_machine_init(void) {
    current_state = STATE_IDLE;
    ESP_LOGI(TAG, "state machine initialized, state=idle");
}

device_state_t state_machine_get_state(void) {
    return current_state;
}

const char *state_machine_get_state_name(void) {
    return state_names[current_state];
}

int state_machine_transition(device_state_t new_state) {
    if (new_state == current_state) return 0;

    // validate transitions
    switch (current_state) {
    case STATE_IDLE:
        if (new_state != STATE_HEATING) return -1;
        break;
    case STATE_HEATING:
        if (new_state != STATE_RUNNING && new_state != STATE_ERROR) return -1;
        break;
    case STATE_RUNNING:
        if (new_state != STATE_DONE && new_state != STATE_ERROR) return -1;
        break;
    case STATE_DONE:
        if (new_state != STATE_IDLE) return -1;
        break;
    case STATE_ERROR:
        if (new_state != STATE_IDLE) return -1;
        break;
    }

    device_state_t old = current_state;
    current_state = new_state;
    ESP_LOGI(TAG, "state transition: %s -> %s", state_names[old], state_names[new_state]);

    if (callback) callback(old, new_state);
    return 0;
}

void state_machine_set_callback(state_change_callback_t cb) {
    callback = cb;
}
