package correlation

import (
	"context"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gitlab.com/feedplan-libraries/common/constants"
)

// WithReqContext returns logger
func WithReqContext(c *gin.Context) context.Context {
	correlationId := c.GetHeader(constants.CorrelationId)
	if len(correlationId) == 0 {
		correlationID, _ := uuid.NewUUID()
		correlationId = correlationID.String()
		c.Request.Header.Set(constants.CorrelationId, correlationId)
	}
	c.Writer.Header().Set(constants.CorrelationId, correlationId)

	requestCtx := context.WithValue(context.Background(), constants.CorrelationId, correlationId)
	return requestCtx
}

func ContextCorrelationId(ctx context.Context) string {
	if ctxCorrelationID, ok := ctx.Value(constants.CorrelationId).(string); ok {
		return ctxCorrelationID
	}
	return ""
}
