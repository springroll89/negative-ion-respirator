#ifndef STATE_MACHINE_H
#define STATE_MACHINE_H

typedef enum {
    STATE_IDLE,
    STATE_HEATING,
    STATE_RUNNING,
    STATE_DONE,
    STATE_ERROR,
} device_state_t;

typedef void (*state_change_callback_t)(device_state_t old_state, device_state_t new_state);

void state_machine_init(void);
device_state_t state_machine_get_state(void);
const char *state_machine_get_state_name(void);
int state_machine_transition(device_state_t new_state);
void state_machine_set_callback(state_change_callback_t cb);

#endif
