package main

import (
	"bytes"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/joho/godotenv"
	"github.com/sirupsen/logrus"
	"net/http"
	"os"
	"time"
)

var log = logrus.New()

func init() {
	// Load environment. variables from .env file
	_ = godotenv.Load()
}

var (
	room = make(map[string]*websocket.Conn)
)

func WebSocket(c *gin.Context) {
	var upgrader = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin: func(r *http.Request) bool {
			//logger.Infof("升级协议:%s", r.Header["User-Agent"])
			return true
		},
	}

	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.WithField("func", "InitWebSocket").WithError(err).Warn("websocket upgrade error")
		return
	}
	room[conn.RemoteAddr().String()] = conn
	defer conn.Close()

	for {
		switch messageType, p, err := conn.ReadMessage(); {
		case websocket.IsCloseError(err, websocket.CloseNormalClosure):
			return
		case err != nil:
			return
		case messageType == websocket.TextMessage:
			log.Debugf("receive message:%s", p)
			for s, w := range room {
				err := w.WriteMessage(messageType, p)
				if err != nil {
					delete(room, s)
					log.Warnf("write message error:%v", err)
				}
			}
		case messageType == websocket.BinaryMessage:
			for s, w := range room {
				err := w.WriteMessage(messageType, p)
				if err != nil {
					delete(room, s)
					log.Warnf("write message error:%v", err)
				}
			}
		case messageType == websocket.PingMessage:
			err := conn.WriteMessage(websocket.PongMessage, []byte(""))
			if err != nil {
				log.Warnf("write pong message error:%v", err)
			}
		default:

		}
	}

}

func main() {
	// Get configuration from environment variables
	listen := os.Getenv("LISTEN")
	if listen == "" {
		listen = ":10000"
	}

	// Initialize Gin router
	r := gin.Default()

	r.NoRoute(func(c *gin.Context) {
		c.Header("Content-Type", "text/event-stream")
		c.Header("Cache-Control", "no-cache")
		c.Header("Connection", "keep-alive")
		var words = []string{"what", "can", "I", "do", "for", "you?", "what can I do for you?"}
		for _, word := range words {
			c.SSEvent("message", word)
			c.Writer.Flush()
			time.Sleep(1 * time.Second)
		}
	})

	r.Use(Cors())

	r.Any("/", func(c *gin.Context) {
		var buf bytes.Buffer
		for s, strings := range c.Request.Header {
			buf.WriteString(fmt.Sprintf("%s: %v\n", s, strings))
		}
		buf.WriteString(fmt.Sprintf("%s: %v\n", "Host", c.Request.Host))
		buf.WriteString(c.RemoteIP())
		c.String(http.StatusOK, buf.String())
	})

	//204
	r.GET("/generate_204", func(c *gin.Context) {
		c.Status(http.StatusNoContent)
	})

	r.GET("/ws", WebSocket)

	r.GET("/sse", func(c *gin.Context) {
		c.Header("Content-Type", "text/event-stream")
		c.Header("Cache-Control", "no-cache")
		c.Header("Connection", "keep-alive")
		var words = []string{"what", "can", "I", "do", "for", "you?", "what can I do for you?"}
		for _, word := range words {
			c.SSEvent("message", word)
			c.Writer.Flush()
			time.Sleep(1 * time.Second)
		}
	})

	// Start the server
	logrus.Infof("Starting server %s", listen)
	if err := r.Run(listen); err != nil {
		logrus.Fatalf("Failed to start server: %v", err)
	}
}

func Cors() gin.HandlerFunc {

	return func(c *gin.Context) {
		method := c.Request.Method
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE, UPDATE")
		c.Header("Access-Control-Allow-Headers", "Origin, X-Requested-With, Content-Type, Accept, Authorization")
		c.Header("Access-Control-Allow-Credentials", "true")
		c.Set("content-type", "application/json")
		if method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
		}
		c.Next()
	}
}
