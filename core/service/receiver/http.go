package receiver

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/ice-coldbell/analyze-server/core/model"
	"github.com/ice-coldbell/analyze-server/pkg/errorx"
	"github.com/ice-coldbell/analyze-server/pkg/logger"
	"go.uber.org/zap/zapcore"
)

func (c *core) httpReceiver(cfg *httpConfig) {
	handler := gin.New()
	// TODO : Request log
	handler.POST(cfg.Path, ginRecovery(c.l), c.setRequsetLogger(), c.handle())

	srv := &http.Server{Addr: ":" + cfg.Port, Handler: handler}
	go func() {
		c.l.Info("listen...")
		if err := srv.ListenAndServe(); !errorx.Is(err, http.ErrServerClosed) {
			c.l.WithError(errorx.Wrap(err)).Error("init http receiver")
			return
		}
		c.l.Info("stopped serving new connection")
	}()

	c.addStopFunction("http receiver", func() error {
		c.l.Debug("start shutdown...")
		ctx, cancel := context.WithTimeout(
			context.Background(),
			time.Duration(cfg.ShutdownTimeoutSec)*time.Second,
		)
		defer cancel()
		if err := srv.Shutdown(ctx); err != nil {
			return errorx.Wrap(err)
		}
		c.l.Info("graceful shutdown complete")
		return nil
	})
}

func (c *core) handle() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var body requestBody
		if err := ctx.ShouldBind(&body); err != nil {
			c.l.WithError(errorx.Wrap(err)).Info("bad request", logger.Object("request_body", body))
			ctx.AbortWithStatus(http.StatusBadRequest)
			return
		}
		c.enqueueEvent(body.toEvent())
	}
}

type requestBody struct {
	Identifier *string         `json:"identifier" binding:"required"`
	UserID     *string         `json:"user_id,omitempty"`
	EventData  json.RawMessage `json:"data,omitempty"`
}

func (rb requestBody) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	if rb.Identifier != nil {
		logger.Stringp("identifier", rb.Identifier).AddTo(enc)
	}
	if rb.UserID != nil {
		logger.Stringp("user_id", rb.UserID).AddTo(enc)
	}
	if len(rb.EventData) != 0 {
		logger.ByteString("event_data", rb.EventData).AddTo(enc)
	}
	return nil
}

func (rb *requestBody) toEvent() model.Event {
	eventType := model.EventTypeNone
	if rb.UserID != nil {
		eventType = model.EventTypeUser
	}
	return model.NewEvent(eventType, *rb.Identifier, *rb.UserID, rb.EventData)
}
