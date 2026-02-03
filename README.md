# Button Event Hub & IoT Tracker

A real-time monitoring system for wireless physical buttons using an **ESP32** hardware layer, a **Go (Golang)** backend, and a **WebSocket-based** web dashboard.

[Image of an IoT architecture diagram showing UDP discovery, HTTP POST, and WebSocket communication]

## 🚀 System Overview

The system is designed for high reliability and zero-configuration setup. It uses a decentralized discovery model so that hardware units can find the central server automatically without manual IP entry.

- **Dynamic Discovery:** Hardware finds the server via UDP broadcast on port **50090**.
- **Real-Time Sync:** Events are pushed from the Go server to the dashboard via WebSockets for sub-second latency.
- **Minimalist Hardware:** Optimized for the ESP32 using a single GPIO pin.

---

## 🛠 Technical Specifications

### 1. Backend (Go)

- **HTTP Port**: `3000`.
- **UDP Discovery Port**: `50090`.
- **Endpoints**:
  - `POST /button`: Receives hardware events via JSON.
  - `GET /ws`: WebSocket stream for live web clients.
  - `GET /`: Serves the dashboard UI via `dashHandler`.
- **Asset Management**: Serves files from the relative directory `../public/index.html`.

### 2. Hardware (ESP32)

- **Microcontroller**: ESP32 (Wi-Fi enabled).
- **Button Pin**: `GPIO 2`.
- **Operation**: On press (LOW signal), the device resolves the server via UDP and sends an HTTP POST.

---

## 📡 Communication Protocols

### Discovery Handshake

The ESP32 discovers the Go server using the following flow:

1.  **Hardware** broadcasts `"DISCOVER_BUTTON_HUB"` to UDP port **50090**.
2.  **Server** responds with a formatted string: `BUTTON_HUB|<IP_ADDRESS>|3000`.
3.  **Hardware** parses this string to construct the dynamic URL: `http://<IP>:3000/button`.

### Hardware Data Payload (JSON)

Hardware events are sent to the `/button` endpoint as follows:

```json
{
  "color": "red"
}
```

_Note: The server automatically generates a nanosecond-precision timestamp for the event upon receipt._

### WebSocket Broadcast (JSON)

The dashboard receives the processed event containing the server-side timestamp:

```json
{
  "color": "red",
  "timestamp": 1706981234000
}
```

---

## 📁 Repository Structure

```text
.
├── server/
│   ├── main.go         # Go server with Discovery & WebSockets
│   ├── go.sum
│   └── go.mod
├── button-code.cpp     # ESP32 Arduino C++ firmware (BTN_PIN 2)
└── public/
    └── index.html      # Live Dashboard UI
```

---

## 🚦 Getting Started

### 1. Setup the Backend

Ensure you have the `gorilla/websocket` package installed:

```bash
go get github.com/gorilla/websocket
cd backend
go run main.go
```

### 2. Flash the Firmware

1.  Open the Arduino code in the Arduino IDE.
2.  Enter your Wi-Fi credentials (`ssid` and `password`).
3.  Ensure `BTN_PIN` is set to `2`.
4.  Select **ESP32 Dev Module** and upload.

### 3. Open the Dashboard

Navigate to `http://localhost:3000` to see live events as buttons are pressed.
