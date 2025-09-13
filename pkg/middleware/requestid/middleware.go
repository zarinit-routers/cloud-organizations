package requestid

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

const Header = "X-Request-ID"

// Middleware ensures each request has a request id in headers and context.
func Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		rid := c.GetHeader(Header)
		if rid == "" {
			rid = uuid.NewString()
		}
		c.Writer.Header().Set(Header, rid)
		c.Set("request_id", rid)
		c.Next()
	}
}
