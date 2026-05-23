#include "heater_ctrl.h"
#include "esp_log.h"
#include "driver/ledc.h"
#include "driver/adc.h"
#include "esp_adc_cal.h"

static const char *TAG = "heater_ctrl";
static int gpio_pwm = 0, gpio_adc_heat = 0, gpio_adc_out = 0;
static int heat_max_temp = 80, target_out_temp = 35;
static float current_heat = 0, current_out = 0;
static uint8_t pwm_duty = 0;
static bool running = false;

// Simple PID coefficients
static float kp = 2.0, ki = 0.1, kd = 0.5;
static float integral = 0, prev_error = 0;
static float output_max = 100.0, output_min = 0.0;
static float integral_max = 50.0;

static float read_temp(int gpio) {
    // Simplified: real implementation uses ADC calibration and NTC thermistor lookup
    // Returns temperature in Celsius
    int raw = adc1_get_raw(gpio == gpio_adc_heat ? ADC1_CHANNEL_0 : ADC1_CHANNEL_3);
    return 25.0 + (raw - 2048) * 0.05; // dummy conversion
}

static float pid_compute(float setpoint, float current) {
    float error = setpoint - current;
    integral += error;
    if (integral > integral_max) integral = integral_max;
    if (integral < -integral_max) integral = -integral_max;

    float derivative = error - prev_error;
    float output = kp * error + ki * integral + kd * derivative;
    if (output > output_max) output = output_max;
    if (output < output_min) output = output_min;

    prev_error = error;
    return output;
}

esp_err_t heater_ctrl_init(int pwm_pin, int adc_heat_pin, int adc_out_pin) {
    gpio_pwm = pwm_pin;
    gpio_adc_heat = adc_heat_pin;
    gpio_adc_out = adc_out_pin;

    // Configure LEDC PWM for heater
    ledc_timer_config_t timer = {
        .speed_mode = LEDC_LOW_SPEED_MODE,
        .duty_resolution = LEDC_TIMER_10_BIT,
        .timer_num = LEDC_TIMER_0,
        .freq_hz = 1000,
        .clk_cfg = LEDC_AUTO_CLK,
    };
    ledc_timer_config(&timer);

    ledc_channel_config_t channel = {
        .gpio_num = gpio_pwm,
        .speed_mode = LEDC_LOW_SPEED_MODE,
        .channel = LEDC_CHANNEL_0,
        .timer_sel = LEDC_TIMER_0,
        .duty = 0,
        .hpoint = 0,
    };
    ledc_channel_config(&channel);

    // Configure ADC
    adc1_config_width(ADC_WIDTH_BIT_12);
    adc1_config_channel_atten(ADC1_CHANNEL_0, ADC_ATTEN_DB_11);
    adc1_config_channel_atten(ADC1_CHANNEL_3, ADC_ATTEN_DB_11);

    ESP_LOGI(TAG, "heater control initialized: pwm=%d adc_heat=%d adc_out=%d",
             pwm_pin, adc_heat_pin, adc_out_pin);
    return ESP_OK;
}

esp_err_t heater_ctrl_set_target(int heat_max, int out_target) {
    heat_max_temp = heat_max;
    target_out_temp = out_target;
    ESP_LOGI(TAG, "target updated: heat_max=%d out_target=%d", heat_max, out_target);
    return ESP_OK;
}

esp_err_t heater_ctrl_get_temps(float *heat_temp, float *out_temp) {
    *heat_temp = current_heat;
    *out_temp = current_out;
    return ESP_OK;
}

esp_err_t heater_ctrl_run(void) {
    if (!running) return ESP_OK;

    current_heat = read_temp(gpio_adc_heat);
    current_out = read_temp(gpio_adc_out);

    // Safety check: hardware over-temp protection
    if (current_heat > 85.0) {
        ESP_LOGE(TAG, "CRITICAL: over temperature detected! heat=%.1f", current_heat);
        heater_ctrl_emergency_shutdown();
        return ESP_FAIL;
    }

    // Software limit: cap heat at max
    if (current_heat > heat_max_temp) {
        pwm_duty = 0;
        integral = 0;
    } else {
        float output = pid_compute(target_out_temp, current_out);
        pwm_duty = (uint8_t)(output * 10.23); // 0-100% -> 0-1023
    }

    ledc_set_duty(LEDC_LOW_SPEED_MODE, LEDC_CHANNEL_0, pwm_duty);
    ledc_update_duty(LEDC_LOW_SPEED_MODE, LEDC_CHANNEL_0);
    return ESP_OK;
}

esp_err_t heater_ctrl_stop(void) {
    running = false;
    pwm_duty = 0;
    integral = 0;
    prev_error = 0;
    ledc_set_duty(LEDC_LOW_SPEED_MODE, LEDC_CHANNEL_0, 0);
    ledc_update_duty(LEDC_LOW_SPEED_MODE, LEDC_CHANNEL_0);
    ESP_LOGI(TAG, "heater stopped");
    return ESP_OK;
}

esp_err_t heater_ctrl_emergency_shutdown(void) {
    heater_ctrl_stop();
    ESP_LOGE(TAG, "EMERGENCY SHUTDOWN executed");
    return ESP_OK;
}
