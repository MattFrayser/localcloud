package api

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

func (s *Server) listContainers(c *gin.Context) {
	containers := s.manager.List()
	c.JSON(http.StatusOK, Response{
		Success: true,
		Data:    containers,
	})
}

func (s *Server) createContainer(c *gin.Context) {
	var req struct {
		Image string `json:"image"`
		Name  string `json:"name"`
		Ports string `json:"ports"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, Response{
			Success: false,
			Error:   "Invalid request format",
		})
		return
	}

	if req.Image == "" {
		req.Image = "nginx:latest"
	}

	instance, err := s.manager.Create(req.Image, req.Name, req.Ports)
	if err != nil {
		c.JSON(http.StatusInternalServerError, Response{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, Response{
		Success: true,
		Data:    instance,
	})
}

func (s *Server) deleteContainer(c *gin.Context) {
	containerID := c.Param("id")
	
	if err := s.manager.Delete(containerID); err != nil {
		c.JSON(http.StatusInternalServerError, Response{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, Response{
		Success: true,
	})
}

func (s *Server) getContainerLogs(c *gin.Context) {
	containerID := c.Param("id")
	tail := 100

	if tailParam := c.Query("tail"); tailParam != "" {
		if parsed, err := strconv.Atoi(tailParam); err == nil {
			tail = parsed
		}
	}

	logs, err := s.manager.GetLogs(containerID, tail)
	if err != nil {
		c.JSON(http.StatusInternalServerError, Response{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, Response{
		Success: true,
		Data:    logs,
	})
}

func (s *Server) getContainerMetrics(c *gin.Context) {
	containerID := c.Param("id")
	
	metrics, err := s.manager.GetMetrics(containerID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, Response{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, Response{
		Success: true,
		Data:    metrics,
	})
}

func (s *Server) execContainer(c *gin.Context) {
	containerID := c.Param("id")
	
	var req struct {
		Command string `json:"command"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, Response{
			Success: false,
			Error:   "Invalid request format",
		})
		return
	}

	output, err := s.manager.Exec(containerID, req.Command)
	if err != nil {
		c.JSON(http.StatusInternalServerError, Response{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, Response{
		Success: true,
		Data:    output,
	})
}
