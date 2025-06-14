package api

import (
	"fmt"
	"log"

	"localcloud/internal/compute"
	"localcloud/internal/config"

	"github.com/gin-gonic/gin"
)

type Server struct {
	manager *compute.Manager
	config  *config.Config
	router  *gin.Engine
}

type Response struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}

func NewServer(manager *compute.Manager, cfg *config.Config) *Server {
	if cfg.LogLevel != "DEBUG" {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.New()
	router.Use(gin.Logger(), gin.Recovery())
	
	s := &Server{
		manager: manager,
		config:  cfg,
		router:  router,
	}

	s.setupRoutes()
	return s
}

func (s *Server) Start() error {
	addr := fmt.Sprintf(":%d", s.config.Port)
	log.Printf("LocalCloud web interface starting on http://localhost%s", addr)
	return s.router.Run(addr)
}

func (s *Server) setupRoutes() {
	// Serve static dashboard
	s.router.GET("/", s.handleDashboard)
	
	// API routes
	api := s.router.Group("/api/v1")
	{
		api.GET("/containers", s.listContainers)
		api.POST("/containers", s.createContainer)
		api.DELETE("/containers/:id", s.deleteContainer)
		api.GET("/containers/:id/logs", s.getContainerLogs)
		api.GET("/containers/:id/metrics", s.getContainerMetrics)
		api.POST("/containers/:id/exec", s.execContainer)
	}

	// WebSocket for real-time updates
	s.router.GET("/ws", s.handleWebSocket)
}
