package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

func AuthBearerKey(token string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// w := c.Writer
		// r := c.Request
		// ctx := r.Context()

		value := c.GetHeader("Authorization")
		if value == "" {
			// oerr := dto.OAuth2ErrorInvalidRequest(http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
			// log.Error(ctx, http.StatusText(http.StatusBadRequest), logger.UrlField(r.URL.String()))
			// http.Error(w, oerr.Error(), oerr.GetStatusCode())
			c.JSON(http.StatusBadRequest, gin.H{"error": "Authorization header is required"})
			c.Abort()
			return
		}

		arr := strings.Split(value, " ")
		if len(arr) != 2 || !strings.EqualFold(arr[0], "Bearer") {
			// oerr := dto.OAuth2ErrorInvalidRequest(http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
			// log.Error(ctx, http.StatusText(http.StatusBadRequest), logger.UrlField(r.URL.String()))
			// http.Error(w, oerr.Error(), oerr.GetStatusCode())
			c.JSON(http.StatusBadRequest, gin.H{"error": "Authorization header is invalid"})
			c.Abort()
			return
		}

		if arr[1] != token {
			// oerr := dto.OAuth2ErrorInvalidRequest(http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
			// log.Error(ctx, http.StatusText(http.StatusUnauthorized), logger.UrlField(r.URL.String()))
			// http.Error(w, oerr.Error(), oerr.GetStatusCode())
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization key is invalid"})
			c.Abort()
			return
		}

		c.Next()
	}
}
