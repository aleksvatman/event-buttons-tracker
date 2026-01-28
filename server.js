const express = require('express');
const WebSocket = require('ws');
const cors = require('cors');
const path = require('path');
const { Bonjour } = require('bonjour-service');

const app = express();
const PORT = 3000;

// Middleware
app.use(cors());
app.use(express.json());
app.use(express.static('public')); // Serve static files from public directory

// Create HTTP server
const server = require('http').createServer(app);

// Create WebSocket server
const wss = new WebSocket.Server({ server });

// Store connected clients
const clients = new Set();

// WebSocket connection handling
wss.on('connection', (ws, req) => {
  const clientId = Math.random().toString(36).substring(7);
  clients.add(ws);

  console.log(`🔌 New WebSocket client connected (ID: ${clientId}). Total clients: ${clients.size}`);

  // Send welcome message
  ws.send(JSON.stringify({
    type: 'connection',
    message: 'Connected to button tracker server',
    clientId: clientId,
    timestamp: new Date().toISOString()
  }));

  // Handle client disconnect
  ws.on('close', () => {
    clients.delete(ws);
    console.log(`🔌 Client disconnected (ID: ${clientId}). Total clients: ${clients.size}`);
  });

  // Handle WebSocket errors
  ws.on('error', (error) => {
    console.error(`❌ WebSocket error for client ${clientId}:`, error);
    clients.delete(ws);
  });

  // Handle ping/pong for connection health
  ws.on('pong', () => {
    ws.isAlive = true;
  });
});

// Ping clients periodically to check connection health
const interval = setInterval(() => {
  wss.clients.forEach((ws) => {
    if (ws.isAlive === false) {
      clients.delete(ws);
      return ws.terminate();
    }

    ws.isAlive = false;
    ws.ping();
  });
}, 30000);

// Clean up interval on server close
wss.on('close', () => {
  clearInterval(interval);
});

// Function to broadcast message to all connected clients
function broadcastToClients(message) {
  const messageStr = JSON.stringify(message);
  let successCount = 0;

  clients.forEach((client) => {
    if (client.readyState === WebSocket.OPEN) {
      try {
        client.send(messageStr);
        successCount++;
      } catch (error) {
        console.error('❌ Error sending message to client:', error);
        clients.delete(client);
      }
    } else {
      // Remove closed connections
      clients.delete(client);
    }
  });

  console.log(`📡 Broadcasted message to ${successCount} clients`);
  return successCount;
}

// Routes
app.get('/', (req, res) => {
  res.sendFile(path.join(__dirname, 'public', 'index.html'));
});

app.get('/status', (req, res) => {
  res.json({
    status: 'running',
    connectedClients: clients.size,
    uptime: process.uptime(),
    timestamp: new Date().toISOString()
  });
});

// Main button endpoint - receives POST requests from Arduino
app.post('/button', (req, res) => {
  const { color } = req.body;

  // Validate request
  if (!color) {
    return res.status(400).json({
      error: 'Missing color',
      expected: '{ "color": "red" }'
    });
  }

  // Send immediate response to Arduino (fast!)
  res.json({
    success: true,
    color: color
  });

  // Process event after responding (non-blocking)
  setImmediate(() => {
    const timestamp = new Date().toISOString();
    const buttonEvent = {
      type: 'button_press',
      color: color,
      timestamp: timestamp,
      id: Math.random().toString(36).substring(7)
    };

    console.log(`🔴 Button pressed: ${color} at ${timestamp}`);

    // Broadcast to WebSocket clients
    const clientCount = broadcastToClients(buttonEvent);
    console.log(`📡 Broadcasted to ${clientCount} clients`);
  });
});

// Error handling middleware
app.use((err, req, res, next) => {
  console.error('❌ Server error:', err.stack);
  res.status(500).json({ error: 'Internal server error' });
});

// 404 handler
app.use((req, res) => {
  res.status(404).json({ error: 'Endpoint not found' });
});

// Optimize server for faster responses
server.keepAliveTimeout = 5000;
server.headersTimeout = 6000;

// Start server
server.listen(PORT, '0.0.0.0', () => {
  console.log(`🚀 Button Tracker Server running on port ${PORT}`);
  console.log(`📡 WebSocket server ready for connections`);

  // Enable mDNS advertising with Bonjour
  const bonjour = new Bonjour();

  try {
    console.log('🔧 Setting up mDNS service...');

    const service = bonjour.publish({
      name: 'button-tracker',
      type: 'http',
      port: PORT,
      host: 'button-tracker.local',
      txt: {
        path: '/button',
        version: '1.0.0'
      }
    });

    service.on('up', () => {
      console.log(`✅ mDNS service published: button-tracker.local:${PORT}`);
      console.log(`🔗 Arduino endpoint: http://button-tracker.local:${PORT}/button`);
      console.log(`📊 Dashboard: http://button-tracker.local:${PORT}`);
      console.log(`📈 Status: http://button-tracker.local:${PORT}/status`);
    });

    service.on('error', (err) => {
      console.log(`❌ mDNS error: ${err.message}`);
      console.log(`🔧 Trying alternative mDNS setup...`);
    });

    // Give it time to register
    setTimeout(() => {
      console.log('🧪 Testing mDNS registration...');
    }, 2000);

  } catch (error) {
    console.log(`⚠️  mDNS not available: ${error.message}`);
    console.log(`📋 Server still accessible via IP address`);
  }
});

// Graceful shutdown
process.on('SIGTERM', () => {
  console.log('🛑 SIGTERM received, shutting down gracefully');
  server.close(() => {
    console.log('✅ Server closed');
    process.exit(0);
  });
});

process.on('SIGINT', () => {
  console.log('🛑 SIGINT received, shutting down gracefully');
  server.close(() => {
    console.log('✅ Server closed');
    process.exit(0);
  });
});