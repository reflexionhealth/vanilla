package httpbase

import (
	"bytes"
	"io"
	"log"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/reflexionhealth/vanilla/router"
)

const (
	AnsiBlack   = "\x1b[30m"
	AnsiRed     = "\x1b[31m"
	AnsiGreen   = "\x1b[32m"
	AnsiYellow  = "\x1b[33m"
	AnsiBlue    = "\x1b[34m"
	AnsiMagenta = "\x1b[35m"
	AnsiCyan    = "\x1b[36m"
	AnsiWhite   = "\x1b[37m"

	AnsiReset = "\x1b[0m"
	AnsiBold  = "\x1b[1m"

	LogTimeFormat = "2006/01/02 - 15:04:05"
)

var Logger = NewRflxLogger(os.Stdout)

// RflxLogger stores log output in memory for a given request context so that log
// output for the given request is sequential in the final log.
// This makes it easier to gobble up all the information for a single request with Logstash.
type RflxLogger struct {
	Global *log.Logger
	Pool   sync.Pool
}

func NewRflxLogger(out io.Writer) *RflxLogger {
	logger := &RflxLogger{log.New(out, "", 0), sync.Pool{}}
	logger.Pool.New = newRequestLog
	return logger
}

type RequestLog struct {
	*log.Logger
	Buffer *bytes.Buffer
}

func newRequestLog() interface{} {
	buffer := &bytes.Buffer{}
	return &RequestLog{log.New(buffer, "", 0), buffer}
}

func (l *RflxLogger) Logf(c *router.Context, format string, args ...interface{}) {
	logPtr, exists := c.GetLocal("Log")
	if exists {
		logger := logPtr.(*RequestLog)
		logger.Printf(format, args...)
	} else {
		Logger.Global.Printf(format, args...)
	}
}

func (l *RflxLogger) LogValue(c *router.Context, name string, value interface{}) {
	logPtr, exists := c.GetLocal("Log")
	if exists {
		logger := logPtr.(*RequestLog)
		logger.Printf(" -- %s%s:%s %v\n", AnsiBold, name, AnsiReset, value)
	} else {
		// LogValue should only be called after the LogRequest middleware,
		// Print out a [?] if we don't have a "Log" local
		Logger.Global.Printf("[?] %s%s:%s %v\n", AnsiBold, name, AnsiReset, value)
	}
}

func (l *RflxLogger) LogResponse(c *router.Context, status string, value interface{}) {
	logPtr, exists := c.GetLocal("Log")
	if exists {
		logger := logPtr.(*RequestLog)
		logger.Printf(" -> %s%s:%s %v\n", AnsiBold, status, AnsiReset, value)
	} else {
		// LogValue should only be called after the LogRequest middleware,
		// Print out a [?] if we don't have a "Log" local
		Logger.Global.Printf("[?] %s%s:%s %v\n", AnsiBold, status, AnsiReset, value)
	}
}

func LogRequest(c *router.Context) {
	start := time.Now()
	path := c.Request.URL.Path
	method := c.Request.Method
	clientIP := c.ClientIP()

	// Always immediately log that we received a request, in case the request takes a long time
	Logger.Global.Printf("Received %s \"%s\" from %s at %v\n", method, path, clientIP, start.Format(LogTimeFormat))

	// Log preamble
	request := Logger.Pool.Get().(*RequestLog)
	request.Buffer.Reset()
	request.Printf("Log for %s \"%s\" from %s at %v\n", method, path, clientIP, start.Format(LogTimeFormat))
	defer Logger.Pool.Put(request)

	c.SetLocal("Log", request)
	defer delete(c.Locals, "Log")

	// Log headers we care about
	headers := []string{"Accept", "Reflexion-Application"}
	for _, header := range headers {
		value := c.Request.Header.Get(header)
		if len(value) > 0 {
			if len(value) <= 60 {
				Logger.LogValue(c, header, value)
			} else {
				Logger.LogValue(c, header, value[:56]+" ...")
			}
		}
	}

	// Handle request
	c.MustContinue() // only use MustContinue for performance critical middleware

	// Log postambole
	end := time.Now()
	latency := end.Sub(start)
	statusCode := c.Response.Status()
	statusText := http.StatusText(statusCode)
	statusColor := colorForStatus(statusCode)
	request.Printf("Replied with %s%d %s%s in %v\n", statusColor, statusCode, statusText, AnsiReset, latency)

	// Write log
	Logger.Global.Print(request.Buffer.String())
}

func colorForStatus(code int) string {
	switch {
	case code >= 100 && code < 200:
		return AnsiBlue
	case code >= 200 && code < 300:
		return AnsiGreen
	case code >= 300 && code < 400:
		return AnsiGreen
	case code >= 400 && code < 500:
		return AnsiYellow
	default:
		return AnsiRed
	}
}
