package middleware

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
)

type bodyLogWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

func (w bodyLogWriter) Write(b []byte) (int, error) {
	w.body.Write(b)
	return w.ResponseWriter.Write(b)
}

func Logger(logger zerolog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		query := c.Request.URL.RawQuery

		var requestBody []byte
		if c.Request.Body != nil {
			var err error
			requestBody, err = io.ReadAll(c.Request.Body)
			if err != nil {
				logger.Warn().Err(err).Msg("failed to read request body")
			}
			c.Request.Body = io.NopCloser(bytes.NewBuffer(requestBody))
		}

		blw := &bodyLogWriter{
			body:           bytes.NewBufferString(""),
			ResponseWriter: c.Writer,
		}
		c.Writer = blw

		c.Next()

		end := time.Now()
		latency := end.Sub(start)

		statusCode := c.Writer.Status()
		clientIP := c.ClientIP()
		method := c.Request.Method
		userAgent := c.Request.UserAgent()
		errorMessage := c.Errors.ByType(gin.ErrorTypePrivate).String()

		responseBody := blw.body.String()

		event := logger.Info()
		if statusCode >= 400 {
			event = logger.Error()
		}

		fields := map[string]interface{}{
			"status":     statusCode,
			"latency":    latency.String(),
			"client_ip":  clientIP,
			"method":     method,
			"path":       path,
			"query":      query,
			"user_agent": userAgent,
		}

		if len(requestBody) > 0 && len(requestBody) < 1024 {
			var prettyJSON bytes.Buffer
			if err := json.Indent(&prettyJSON, requestBody, "", "  "); err == nil {
				fields["request_body"] = prettyJSON.String()
			} else {
				fields["request_body"] = string(requestBody)
			}
		}

		if len(responseBody) > 0 && len(responseBody) < 1024 {
			var prettyJSON bytes.Buffer
			if err := json.Indent(&prettyJSON, []byte(responseBody), "", "  "); err == nil {
				fields["response_body"] = prettyJSON.String()
			} else {
				fields["response_body"] = responseBody
			}
		}

		if errorMessage != "" {
			fields["error"] = errorMessage
		}

		event.Fields(fields).Msg(fmt.Sprintf("%s %s", method, path))
	}
}

func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		requestID := c.Request.Header.Get("X-Request-ID")
		if requestID == "" {
			requestID = generateRequestID()
		}
		c.Set("request_id", requestID)
		c.Writer.Header().Set("X-Request-ID", requestID)
		c.Next()
	}
}

func generateRequestID() string {
	return fmt.Sprintf("%d", time.Now().UnixNano())
}
