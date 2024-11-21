package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func NoCache() gin.HandlerFunc {
	return func(c *gin.Context) {
		setNoCache(c.Writer)
		c.Next()
	}
}

func setNoCache(w http.ResponseWriter) {
	w.Header().Set("Cache-Control", "max-age=0, no-cache, no-store, must-revalidate;")
	w.Header().Set("Pragma", "no-cache")
	w.Header().Set("Expires", "0")
	// w.Header().Set("X-Content-Type-Options", "nosniff")
}
