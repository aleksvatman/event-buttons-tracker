# Button Tracker System

A real-time button tracking system that connects Arduino ESP32 devices to a web dashboard via WebSocket communication.

## 🏗️ System Architecture

```
Arduino ESP32 → HTTP POST → Node.js Server → WebSocket → Web Dashboard
```

## 📁 Project Structure

```
button-tracker/
├── server.js           # Node.js Express + WebSocket server
├── package.json        # Dependencies and scripts
├── button-code.cpp     # Arduino ESP32 code
├── public/
│   └── index.html      # Web dashboard client
└── README.md           # This file
```

## 🚀 Quick Start

### 1. Install Dependencies

```bash
npm install
```

### 2. Start the Server

```bash
npm start
# or for development with auto-restart:
npm run dev
```

### 3. Open Web Dashboard

Navigate to: http://localhost:3000

### 4. Configure Arduino

1. Update `button-code.cpp` with your WiFi credentials:

   ```cpp
   const char* ssid = "YOUR_WIFI_NAME";
   const char* password = "YOUR_WIFI_PASSWORD";
   ```

2. Update the server URL with your computer's IP:

   ```cpp
   const char* serverURL = "http://YOUR_PC_IP:3000/button";
   ```

3. Set the button color:
   ```cpp
   const char* buttonColor = "red"; // red, green, blue, yellow, white
   ```

### 5. Upload to ESP32

Upload the code to your ESP32 using Arduino IDE.

## 🔌 API Endpoints

### POST /button

Receives button press events from Arduino devices.

**Request Body:**

```json
{
  "color": "red"
}
```

**Response:**

```json
{
  "success": true,
  "message": "Button press (red) received and broadcasted",
  "clientsNotified": 2,
  "timestamp": "2024-01-15T10:30:00.000Z"
}
```

### GET /status

Returns server status and connected client count.

**Response:**

```json
{
  "status": "running",
  "connectedClients": 2,
  "uptime": 1234.56,
  "timestamp": "2024-01-15T10:30:00.000Z"
}
```

### GET /

Serves the web dashboard interface.

## 🌐 WebSocket Events

### From Server to Client

**Connection Event:**

```json
{
  "type": "connection",
  "message": "Connected to button tracker server",
  "clientId": "abc123",
  "timestamp": "2024-01-15T10:30:00.000Z"
}
```

**Button Press Event:**

```json
{
  "type": "button_press",
  "color": "red",
  "timestamp": "2024-01-15T10:30:00.000Z",
  "id": "def456"
}
```

## 🔧 Hardware Setup

### Components Needed

- ESP32 development board
- Push button
- 10kΩ resistor (optional, using internal pull-up)
- Breadboard and jumper wires

### Wiring

```
ESP32 Pin 2 → Button → GND
```

The code uses `INPUT_PULLUP`, so no external resistor is needed.

## 🎨 Web Dashboard Features

- **Real-time Connection Status**: Shows WebSocket connection state
- **Live Button Display**: Visual representation of the last button pressed
- **Activity Log**: Chronological list of recent button presses
- **Statistics**: Counters for different button colors
- **Responsive Design**: Works on desktop and mobile devices
- **Auto-reconnection**: Automatically reconnects if WebSocket connection drops

## 🔧 Configuration

### Server Configuration

Edit `server.js` to change:

- Port number (default: 3000)
- CORS settings
- WebSocket ping interval
- Maximum client connections

### Arduino Configuration

Edit `button-code.cpp` to change:

- WiFi credentials
- Server URL
- Button pin
- Button color
- Debounce delay

## 🐛 Troubleshooting

### Arduino Not Connecting to WiFi

1. Check WiFi credentials
2. Ensure 2.4GHz network (ESP32 doesn't support 5GHz)
3. Check signal strength

### Server Not Receiving Button Presses

1. Verify computer's IP address: `ipconfig` (Windows) or `ifconfig` (Mac/Linux)
2. Check firewall settings
3. Ensure server is running on port 3000
4. Check Arduino serial monitor for HTTP response codes

### WebSocket Connection Issues

1. Check browser console for errors
2. Verify server is running
3. Try refreshing the page
4. Check for browser WebSocket support

### Finding Your Computer's IP Address

**Windows:**

```cmd
ipconfig
```

**Mac/Linux:**

```bash
ifconfig
# or
ip addr show
```

Look for your local network IP (usually starts with 192.168.x.x or 10.x.x.x)

## 📊 Monitoring and Logs

The server provides comprehensive logging:

- WebSocket connections/disconnections
- Button press events
- HTTP request/response details
- Error handling and client management

Monitor the console output when running the server to see real-time activity.

## 🔄 Multiple Buttons

To use multiple buttons:

1. Create separate Arduino devices with different `buttonColor` values
2. Each device sends to the same server endpoint
3. The web dashboard automatically handles multiple colors
4. Statistics are tracked separately for each color

## 📈 Extending the System

### Adding New Features

- Database storage for button press history
- User authentication and multiple rooms
- Mobile app integration
- REST API for historical data
- Email/SMS notifications
- Integration with other IoT platforms

### Scaling

- Use a production WebSocket library like Socket.IO
- Add Redis for session management
- Implement load balancing for multiple server instances
- Add database persistence

## 📄 License

MIT License - feel free to modify and distribute as needed.
