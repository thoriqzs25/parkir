package health

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/thoriqzs/PARKIR/backend/internal/response"
)

var startTime = time.Now()

func RegisterRoutes(r *gin.Engine, pool *pgxpool.Pool) {
	r.GET("/health", func(c *gin.Context) {
		response.OK(c, gin.H{"status": "ok"})
	})

	r.GET("/health/ready", func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(c.Request.Context(), 2*time.Second)
		defer cancel()

		var one int
		err := pool.QueryRow(ctx, "SELECT 1").Scan(&one)
		if err != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{
				"status":   "error",
				"database": "disconnected",
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"status":   "ok",
			"database": "connected",
		})
	})
}

func RegisterComponentRoutes(r *gin.Engine, pool *pgxpool.Pool) {
	r.GET("/health/components", func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(c.Request.Context(), 2*time.Second)
		defer cancel()

		dbStatus := "connected"
		var one int
		if err := pool.QueryRow(ctx, "SELECT 1").Scan(&one); err != nil {
			dbStatus = "disconnected"
		}

		response.OK(c, gin.H{
			"status": "ok",
			"components": gin.H{
				"api": gin.H{
					"status":        "up",
					"uptime_seconds": int(time.Since(startTime).Seconds()),
				},
				"database": gin.H{
					"status": dbStatus,
				},
			},
			"last_check": time.Now().UTC(),
		})
	})
}
