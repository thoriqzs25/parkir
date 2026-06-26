package health

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

func RegisterRoutes(r *gin.Engine, pool *pgxpool.Pool) {
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status": "ok",
		})
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
