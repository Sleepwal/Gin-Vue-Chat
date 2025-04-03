package websocket

import (
	"log"
	"sync"
)

// Client 是一个中间人，在websocket连接和hub之间
type Client struct {
	Hub  *Hub
	Conn *Connection
	// 用户ID
	UserID string
	// 发送消息的通道
	Send chan []byte
	// 互斥锁，保护连接
	mu sync.Mutex
}

// Hub 维护活跃客户端的集合并广播消息
type Hub struct {
	// 注册的客户端
	clients map[*Client]bool

	// 用户ID到客户端的映射
	userClients map[string]*Client

	// 从客户端入站的消息
	broadcast chan []byte

	// 注册请求
	register chan *Client

	// 注销请求
	unregister chan *Client

	// 互斥锁，保护maps
	mu sync.RWMutex
}

// NewHub 创建一个新的Hub
func NewHub() *Hub {
	return &Hub{
		broadcast:   make(chan []byte),
		register:    make(chan *Client),
		unregister:  make(chan *Client),
		clients:     make(map[*Client]bool),
		userClients: make(map[string]*Client),
		mu:          sync.RWMutex{},
	}
}

// Run 启动hub的消息处理循环
func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.mu.Lock()
			h.clients[client] = true
			if client.UserID != "" {
				h.userClients[client.UserID] = client
				log.Printf("Client registered: %s", client.UserID)
			}
			h.mu.Unlock()

		case client := <-h.unregister:
			h.mu.Lock()
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				if client.UserID != "" {
					delete(h.userClients, client.UserID)
					log.Printf("Client unregistered: %s", client.UserID)
				}
				close(client.Send)
			}
			h.mu.Unlock()

		case message := <-h.broadcast:
			h.mu.RLock()
			for client := range h.clients {
				select {
				case client.Send <- message:
				default:
					h.mu.RUnlock()
					h.mu.Lock()
					delete(h.clients, client)
					if client.UserID != "" {
						delete(h.userClients, client.UserID)
					}
					close(client.Send)
					h.mu.Unlock()
					h.mu.RLock()
				}
			}
			h.mu.RUnlock()
		}
	}
}

// SendToUser 发送消息给特定用户
func (h *Hub) SendToUser(userID string, message []byte) bool {
	h.mu.RLock()
	defer h.mu.RUnlock()

	client, exists := h.userClients[userID]
	if !exists {
		return false
	}

	client.mu.Lock()
	defer client.mu.Unlock()

	select {
	case client.Send <- message:
		return true
	default:
		return false
	}
}

// Broadcast 广播消息给所有连接的客户端
func (h *Hub) Broadcast(message []byte) {
	h.broadcast <- message
}
