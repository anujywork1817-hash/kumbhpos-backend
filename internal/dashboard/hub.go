package dashboard

import (
"encoding/json"
"log"
"net/http"
"sync"

"github.com/gorilla/websocket"
)

type Client struct {
conn *websocket.Conn
send chan []byte
}

type Hub struct {
clients    map[*Client]bool
broadcast  chan []byte
register   chan *Client
unregister chan *Client
mu         sync.Mutex
}

var GlobalHub = NewHub()

func NewHub() *Hub {
return &Hub{
clients:    make(map[*Client]bool),
broadcast:  make(chan []byte, 256),
register:   make(chan *Client),
unregister: make(chan *Client),
}
}

func (h *Hub) Run() {
for {
select {
case client := <-h.register:
h.mu.Lock()
h.clients[client] = true
h.mu.Unlock()
case client := <-h.unregister:
h.mu.Lock()
if _, ok := h.clients[client]; ok {
delete(h.clients, client)
close(client.send)
}
h.mu.Unlock()
case message := <-h.broadcast:
h.mu.Lock()
for client := range h.clients {
select {
case client.send <- message:
default:
close(client.send)
delete(h.clients, client)
}
}
h.mu.Unlock()
}
}
}

func (h *Hub) Broadcast(eventType string, data interface{}) {
payload := map[string]interface{}{
"event": eventType,
"data":  data,
}
b, err := json.Marshal(payload)
if err != nil {
log.Println("broadcast marshal error:", err)
return
}
h.broadcast <- b
}

var upgrader = websocket.Upgrader{
CheckOrigin: func(r *http.Request) bool { return true },
}

func ServeWS(w http.ResponseWriter, r *http.Request) {
conn, err := upgrader.Upgrade(w, r, nil)
if err != nil {
log.Println("ws upgrade error:", err)
return
}
client := &Client{conn: conn, send: make(chan []byte, 256)}
GlobalHub.register <- client

// Write pump
go func() {
defer func() {
GlobalHub.unregister <- client
conn.Close()
}()
for msg := range client.send {
if err := conn.WriteMessage(websocket.TextMessage, msg); err != nil {
return
}
}
}()

// Read pump (keep alive + handle disconnect)
go func() {
defer func() {
GlobalHub.unregister <- client
conn.Close()
}()
for {
if _, _, err := conn.ReadMessage(); err != nil {
return
}
}
}()
}
