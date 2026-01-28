#include <WiFi.h>
#include <HTTPClient.h>
#include <WiFiUdp.h>

const char* ssid = "Vatman 2.4";
const char* password = "bevatman161194";
#define BTN_PIN 2
const char* buttonColor = "red";

String serverURL;

bool lastButtonState = HIGH;
unsigned long lastDebounceTime = 0;
const unsigned long debounceDelay = 50;

bool discoverHub() {
  WiFiUDP udp;
  udp.begin(50091); // клієнтський порт будь-який
  IPAddress broadcastIP(255,255,255,255);
  udp.beginPacket(broadcastIP, 50090);
  udp.write((const uint8_t*)"DISCOVER_BUTTON_HUB", strlen("DISCOVER_BUTTON_HUB"));
  udp.endPacket();

  unsigned long start = millis();
  char buf[128];
  while (millis() - start < 2000) {
    int n = udp.parsePacket();
    if (n) {
      int len = udp.read(buf, 127);
      buf[len] = 0;
      String resp(buf);
      if (resp.startsWith("BUTTON_HUB|")) {
        int p1 = resp.indexOf('|',12);
        String ip = resp.substring(12, p1);
        String port = resp.substring(p1+1);
        serverURL = "http://" + ip + ":" + port + "/button";
        return true;
      }
    }
    delay(50);
  }
  return false;
}

void setup() {
  Serial.begin(115200);
  pinMode(BTN_PIN, INPUT_PULLUP);
  WiFi.mode(WIFI_STA);
  WiFi.begin(ssid, password);
  while (WiFi.status() != WL_CONNECTED) { delay(500); Serial.print("."); }

  bool found = discoverHub();
  if (!found) {
    Serial.println("❌ Hub not found at discovery");
  } else {
    Serial.print("✓ Found hub: "); Serial.println(serverURL);
  }
}

void loop() {
  int reading = digitalRead(BTN_PIN);
  if (reading != lastButtonState) lastDebounceTime = millis();

  if ((millis() - lastDebounceTime) > debounceDelay) {
    if (reading == LOW && lastButtonState == HIGH) {
      sendButtonPress();
    }
  }
  lastButtonState = reading;
}

void sendButtonPress() {
  if (WiFi.status() != WL_CONNECTED || serverURL.length() == 0) return;
  HTTPClient http;
  http.begin(serverURL);
  http.addHeader("Content-Type", "application/json");
  String payload = "{\"color\":\"" + String(buttonColor) + "\"}";
  int code = http.POST(payload);
  Serial.print("Sent! Response: ");
  Serial.println(code);
  http.end();
}
