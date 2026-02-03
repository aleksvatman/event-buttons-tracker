Button Tracking System (Go / React / Arduino)

A complete IoT ecosystem for tracking physical button presses in real-time. This system uses an ESP32 for hardware triggers, a Go (Golang) backend for event processing and discovery, and a WebSocket-enabled web dashboard for live monitoring.
🚀 System Overview

    Automatic Discovery: The hardware finds the server dynamically via a UDP broadcast handshake—no hardcoded IPs required.

    Real-Time Sync: Button events are pushed from the Go server to the Web UI via WebSockets for sub-second latency.

    Persistent State: The server handles incoming hardware triggers and broadcasts them to all connected web clients.

🛠 Technical Specifications

1.  Backend (Go)

    HTTP Port: 3000

    Discovery Port: 50090 (UDP)

    Endpoints:

        POST /button: Receives hardware events.

        GET /ws: WebSocket stream for web clients.

        GET /: Serves the dashboard UI.

    Static Assets: Served from ../public/index.html.

2.  Hardware (ESP32)

    Microcontroller: ESP32 (Wireless)

    Button Pin: GPIO 2 (configured with INPUT_PULLUP).

    Discovery Protocol: Sends "DISCOVER_BUTTON_HUB" to find the server.

📡 Communication Protocol
Discovery Handshake

The ESP32 discovers the Go server using the following flow:

    Hardware broadcasts "DISCOVER_BUTTON_HUB" on UDP port 50090.

    Server responds with a string: BUTTON_HUB|<IP_ADDRESS>|3000.

    Hardware parses the IP and port to construct the dynamic URL: http://<IP>:3000/button.

Data Payload

Hardware events are sent as JSON POST requests:
JSON

```
{
"color": "red"
}
```

The server automatically attaches a timestamp (nanoseconds) upon receipt before broadcasting to web clients.
📁 Repository Structure
Plaintext

```
.
├── server/
│ ├── main.go          # Go server with Discovery & WebSockets
| ├── go.sum
│ └── go.mod
└── button-code.cpp    # ESP32 Arduino C++ firmware (BTN_PIN 2)
└── public/
  └── index.html       # Live Web UI (served by Go)
```

🚦 Getting Started

1. Run the Backend

Ensure you have the gorilla/websocket package installed:
Bash

go get github.com/gorilla/websocket
cd backend
go run main.go

2. Upload Firmware

   Open firmware/button.ino in the Arduino IDE.

   Set your Wi-Fi SSID and Password.

   Ensure BTN_PIN is set to 2.

   Select ESP32 Dev Module and upload.

3. Open Dashboard

Visit http://localhost:3000 in your browser. The UI will connect to the WebSocket and wait for hardware triggers.
