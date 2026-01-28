package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
)

const (
	DISCOVERY_PORT = 50090
	HTTP_PORT      = 3000
)

type ButtonEvent struct {
	Color string `json:"color"`
	Time  int64  `json:"timestamp"`
}

var (
	clients   = make(map[*websocket.Conn]bool)
	clientsMu sync.Mutex
	upgrader  = websocket.Upgrader{}
)

func main() {
	// Start UDP broadcast discovery responder
	go startDiscoveryResponder()

	// HTTP handler for button POSTs
	http.HandleFunc("/button", handleButton)
	// WebSocket handler
	http.HandleFunc("/ws", wsHandler)
	// Dashboard UI
	http.HandleFunc("/", dashHandler)

	fmt.Printf("Button hub: HTTP %d, Discovery %d\n", HTTP_PORT, DISCOVERY_PORT)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", HTTP_PORT), nil))
}

func handleButton(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Only POST allowed", http.StatusMethodNotAllowed)
		return
	}
	var evt ButtonEvent
	if err := json.NewDecoder(r.Body).Decode(&evt); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}
	evt.Time = makeTimestamp()
	broadcastEvent(evt)
	w.WriteHeader(200)
	w.Write([]byte(`{"status":"ok"}`))
}

// --- Discovery responder ---
func startDiscoveryResponder() {
	addr := net.UDPAddr{
		Port: DISCOVERY_PORT,
		IP:   net.ParseIP("0.0.0.0"),
	}
	conn, err := net.ListenUDP("udp", &addr)
	if err != nil {
		log.Fatal("UDP listen error:", err)
	}
	defer conn.Close()
	buf := make([]byte, 256)
	for {
		n, remote, err := conn.ReadFromUDP(buf)
		if err == nil && n > 0 {
			msg := string(buf[:n])
			if msg == "DISCOVER_BUTTON_HUB" {
				ip := getOutboundIP()
				resp := fmt.Sprintf("BUTTON_HUB|%s|%d", ip.String(), HTTP_PORT)
				conn.WriteToUDP([]byte(resp), remote)
			}
		}
	}
}

// -- Helper: Get current machine local IP for response --
func getOutboundIP() net.IP {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err == nil {
		defer conn.Close()
		localAddr := conn.LocalAddr().(*net.UDPAddr)
		return localAddr.IP
	}
	// fallback
	return net.ParseIP("127.0.0.1")
}

// -- WebSocket broadcasting --
func wsHandler(w http.ResponseWriter, r *http.Request) {
	upgrader.CheckOrigin = func(r *http.Request) bool { return true }
	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}
	clientsMu.Lock()
	clients[c] = true
	clientsMu.Unlock()
	defer func() {
		clientsMu.Lock()
		delete(clients, c)
		clientsMu.Unlock()
		c.Close()
	}()
	for {
		if _, _, err := c.NextReader(); err != nil {
			break
		}
	}
}

func broadcastEvent(evt ButtonEvent) {
	msg, _ := json.Marshal(evt)
	clientsMu.Lock()
	defer clientsMu.Unlock()
	for c := range clients {
		c.WriteMessage(websocket.TextMessage, msg)
	}
}

// -- Dashboard minimal UI --
func dashHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte(dashboardHTML))
}

// -- Timestamp helper --
func makeTimestamp() int64 {
	return (int64)(float64(1e3) * float64(float64((float64)(1e-6)*float64(float64((float64)(1e9)*float64(float64(1)))))))
}

const dashboardHTML = `
<!DOCTYPE html>
<html>
<head>
  <title>Live Button Status</title>
  <style>
    .button {
      width: 100px;
      height: 100px;
      border-radius: 50%;
      margin: 20px;
      display: inline-block;
      opacity: 0.3;
      transition: opacity 0.2s;
    }
    .active {
      opacity: 1;
      box-shadow: 0 0 20px black;
    }
    #container {
      text-align: center;
      margin-top: 50px;
    }
  </style>
</head>
<body>
  <h1>Live Button Status</h1>
  <div id="container">
    <div id="red" class="button" style="background-color: red;"></div>
    <div id="green" class="button" style="background-color: green;"></div>
    <div id="blue" class="button" style="background-color: blue;"></div>
  </div>

  <script>
    const ws = new WebSocket("ws://" + location.host + "/ws");
    ws.onmessage = (event) => {
      console.log(event.data);
      const e = JSON.parse(event.data);
      for (const color of ["red", "green", "blue"]) {
        const el = document.getElementById(color);
        if (e.color === color) {
          el.classList.add("active");
        } else {
          el.classList.remove("active");
        }
      }
    };
  </script>
</body>
</html>
`
