package main

import (
	"embed" 
	"encoding/json"
	"fmt"
	"io/fs" 
	"log"
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

//go:embed all:public
var StaticContent embed.FS

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
	go startDiscoveryResponder()

	http.HandleFunc("/button", handleButton)
	http.HandleFunc("/ws", wsHandler)
	http.HandleFunc("/", uiHandler)

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
	evt.Time = time.Now().UnixMilli()
	broadcastEvent(evt)
	w.WriteHeader(200)
	w.Write([]byte(`{"status":"ok"}`))
}

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

func getOutboundIP() net.IP {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err == nil {
		defer conn.Close()
		localAddr := conn.LocalAddr().(*net.UDPAddr)
		return localAddr.IP
	}
	return net.ParseIP("127.0.0.1")
}

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

// -- UI --
func uiHandler(w http.ResponseWriter, r *http.Request) {
	// Merged logic: strips the "public" prefix so we can access "index.html" directly
	publicFS, err := fs.Sub(StaticContent, "public")
	if err != nil {
		log.Printf("FS Sub error: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	data, err := fs.ReadFile(publicFS, "index.html")
	if err != nil {
		log.Printf("Embed error: %v", err)
		http.Error(w, "File not found", http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "text/html")
	w.Write(data)
}