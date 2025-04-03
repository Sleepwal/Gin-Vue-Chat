package websocket

import (
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

const (
	// 允许的写入WebSocket连接的最大时间
	writeWait = 10 * time.Second

	// 允许的读取下一个pong消息的最大时间
	pongWait = 60 * time.Second

	// 发送ping到peer的频率，必须小于pongWait
	pingPeriod = (pongWait * 9) / 10

	// 允许的最大消息大小
	maxMessageSize = 512
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	// 允许所有CORS请求
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

// Connection 封装了websocket连接
type Connection struct {
	// websocket连接
	ws *websocket.Conn

	// 用户ID
	userID string
}

// ServeWs 处理WebSocket请求
func ServeWs(hub *Hub, c *gin.Context) {
	// 从上下文中获取用户ID
	userID, exists := c.Get("userId")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未授权"})
		return
	}

	// 升级HTTP连接为WebSocket连接
	ws, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Println("升级连接失败:", err)
		return
	}

	// 创建连接和客户端
	conn := &Connection{ws: ws, userID: userID.(string)}
	client := &Client{Hub: hub, Conn: conn, UserID: userID.(string), Send: make(chan []byte, 256)}

	// 注册客户端
	client.Hub.register <- client

	// 允许收集未使用的内存
	ws.SetReadLimit(maxMessageSize)
	ws.SetReadDeadline(time.Now().Add(pongWait))
	ws.SetPongHandler(func(string) error { ws.SetReadDeadline(time.Now().Add(pongWait)); return nil })

	// 启动goroutines处理读写
	go client.writePump()
	go client.readPump()
}

// readPump 从WebSocket连接泵取消息
func (c *Client) readPump() {
	defer func() {
		c.Hub.unregister <- c
		c.Conn.ws.Close()
	}()

	for {
		_, message, err := c.Conn.ws.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("错误: %v", err)
			}
			break
		}

		// 处理接收到的消息
		log.Printf("收到消息: %s", message)
		c.Hub.broadcast <- message
	}
}

// writePump 将消息泵送到WebSocket连接
func (c *Client) writePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.Conn.ws.Close()
	}()

	for {
		select {
		case message, ok := <-c.Send:
			c.Conn.ws.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				// Hub关闭了通道
				c.Conn.ws.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.Conn.ws.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			// 添加队列中的所有消息到当前的writer
			n := len(c.Send)
			for i := 0; i < n; i++ {
				w.Write(<-c.Send)
			}

			if err := w.Close(); err != nil {
				return
			}
		case <-ticker.C:
			c.Conn.ws.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.Conn.ws.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}
