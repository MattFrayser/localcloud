// Provide Live conatiner updates 
package api

import (
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

func (s *Server) handleWebSocket(c *gin.Context) {

	// Upgrade HTTP to websocket
	upgrader := websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true // Allow all origins 
		},
	}

	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("WebSocket upgrade error: %v", err)
		return
	}
	defer conn.Close()

	// Send initial data
	containers := s.manager.List()
	initialData := map[string]interface{}{
		"containers": containers,
		"timestamp":  time.Now(),
	}
	
	if err := conn.WriteJSON(initialData); err != nil {
		log.Printf("WebSocket initial write error: %v", err)
		return
	}
	
	// Refresh data every 2 seconds
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			containers := s.manager.List()
			data := map[string]interface{}{
				"containers": containers,
				"timestamp":  time.Now(),
			}

			if err := conn.WriteJSON(data); err != nil {
				log.Printf("WebSocket write error: %v", err)
				return
			}

		case <-c.Request.Context().Done():
			return
		}
	}
}
