package ws

import (
	"net/http"

	"github.com/baking-bad/bcdhub/internal/logger"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

// Handler -
func Handler(c *gin.Context, hub *Hub) {
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		logger.Error(err)
		return
	}

	client := NewClient(conn)
	client.Run()
	hub.AddClient(client)
}
