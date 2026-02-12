# MQTT Topic Structure

## 1. System Summary
Communication is standardized into three distinct channels. All upstream messages (Telemetry and Alerts) must use the standardized JSON envelope.

### Communication Channels
1. **Telemetry (Upstream):** Periodic status updates.
2. **Alerts (Upstream):** Critical safety events sent immediately upon detection.
3. **Commands (Downstream):** Instructions sent to the Device.

---

## 2. Upstream Traffic (Device -> Cloud)

### Standard Envelope
All upstream messages must be wrapped in this structure.

```json
{
  "device_id": "temp-sensor-01",
  "timestamp": 1702588123,
  "type": "sensor",
  "payload": {
      "key": "value"
  }
}
```

### Channel A: Telemetry
* **Topic:** `devices/[device-id]/telemetry`
* **Purpose:** Regular state reporting.

### Channel B: Alerts
* **Topic:** `devices/[device-id]/alerts`
* **Purpose:** Critical events (e.g., Gas Leak).

---

## 3. Downstream Traffic (Cloud -> Device)

### Channel C: Commands
* **Topic:** `devices/[device-id]/command`
* **Payload:** Raw JSON instruction (No envelope required).

**Command Payload Structure:**
```json
{
  "request_id": "req-1",
  "action": "ACTION_NAME",
  "parameters": { "key": "value" }
}
```

---

## 4. Device Dictionary

### 1. Temp Sensor
**ID Pattern**: `temp-sensor-[id]`

**Telemetry Payload:**
```json
{
  "temp": 14.5,
  "status": "COLD" 
}
```

### 2. Light Sensor
**ID Pattern**: `light-sensor-[id]`

**Telemetry Payload:**
```json
{
  "light_level": 450,
  "day_night": "DAY"
}
```

### 3. Gas Sensor
**ID Pattern**: `gas-sensor-[id]`

**Telemetry Payload (Safe):**
```json
{
  "gas_level": 120,
  "status": "SAFE",
  "alarm_on": false
}
```

**Alert Payload (Danger):** *(Send to /alerts topic)*
```json
{
  "gas_level": 950,
  "status": "DANGER",
  "alarm_on": true
}
```

### 4. Door Actuator 
**ID Pattern**: `door-actuator-[id]`

**Telemetry Payload:**
```json
{
  "door_state": "CLOSED",
  "lock_state": "LOCKED"
}
```
**Incoming Commands:**
* Action: `LOCK`
* Action: `UNLOCK`

### 5. A/C Actuator
**ID Pattern**: `ac-actuator-[id]`

**Telemetry Payload:**
```json
{
  "power": "ON",
  "current_temp": 24.0,
  "target_temp": 22.0,
  "mode": "COOLING"
}
```
**Incoming Commands:**
* Action: `SET_STATE`
* Parameters: `power`, `target_temp`, `mode`