package stack

import (
	"bytes"
	"io"
	"log"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/reflexionhealth/vanilla/httpserver"
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

	LogTimeFormat = "2006/01/02 15:04:05"
)

var Logger = NewStackLogger(os.Stdout)

// StackLogger stores log output in memory for a given request context so that log
// output for the given request is sequential in the final log.
// This makes it easier to gobble up all the information for a single request with Logstash.
type StackLogger struct {
	Global *log.Logger
	Pool   sync.Pool
}

func NewStackLogger(out io.Writer) *StackLogger {
	logger := &StackLogger{log.New(out, "", 0), sync.Pool{}}
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

func (l *StackLogger) Logf(c *httpserver.Context, format string, args ...interface{}) {
	logPtr, exists := c.GetLocal("Log")
	if exists {
		logger := logPtr.(*RequestLog)
		logger.Printf(format, args...)
	} else {
		Logger.Global.Printf(format, args...)
	}
}

func (l *StackLogger) LogValue(c *httpserver.Context, name string, value interface{}) {
	logPtr, exists := c.GetLocal("Log")
	if exists {
		logger := logPtr.(*RequestLog)
		if c.Debug {
			logger.Printf(" -- %s%s:%s %v\n", AnsiBold, name, AnsiReset, value)
		} else {
			logger.Printf(" -- %s: %v\n", name, value)
		}
	} else {
		// LogValue should only be called after the LogRequest middleware,
		// Print out a [?] if we don't have a "Log" local
		if c.Debug {
			Logger.Global.Printf("[?] %s%s:%s %v\n", AnsiBold, name, AnsiReset, value)
		} else {
			Logger.Global.Printf("[?] %s: %v\n", name, value)
		}
	}
}

func (l *StackLogger) LogResponse(c *httpserver.Context, status string, value interface{}) {
	logPtr, exists := c.GetLocal("Log")
	if exists {
		logger := logPtr.(*RequestLog)
		if c.Debug {
			logger.Printf(" -> %s%s:%s %v\n", AnsiBold, status, AnsiReset, value)
		} else {
			logger.Printf(" -> %s: %v\n", status, value)
		}
	} else {
		// LogValue should only be called after the LogRequest middleware,
		// Print out a [?] if we don't have a "Log" local
		if c.Debug {
			Logger.Global.Printf("[?] %s%s:%s %v\n", AnsiBold, status, AnsiReset, value)
		} else {
			Logger.Global.Printf("[?] %s: %v\n", status, value)
		}
	}
}

// LogAccess logs a request immediately when a request is received (before it is processed).
// It includes the HTTP method, request path, client IP (w/ X-Forward-For), and time received.
//
// Redundant when used with LogRequest.
// Use LogAccess middleware when either:
//  + Trying to log access to only a subset of requests
//  + Want to log a request without allocating a buffer from the Logger pool
func LogAccess(c *httpserver.Context) {
	start := time.Now()
	path := c.Request.URL.Path
	method := c.Request.Method
	clientIp := c.ClientIp()

	Logger.Global.Printf("Received %s \"%s\" from %s at %v\n", method, path, clientIp, start.Format(LogTimeFormat))

	c.ContinueRequest()
}

// LogError logs a request only if its HTTP status code is greater than 400.
// It includes the HTTP method, request path, client IP (w/ X-Forward-For),
// the error status code and text, the reply time, and the processing latency.
//
// Use the LogError middleware when you don't want the verbosity of LogRequest,
// but still want to log any errors that may occur.
func LogError(c *httpserver.Context) {
	start := time.Now()
	path := c.Request.URL.Path
	method := c.Request.Method
	clientIp := c.ClientIp()

	c.PerformRequest()

	statusCode := c.Response.Status()
	if statusCode >= 400 {
		end := time.Now()
		latency := end.Sub(start)
		statusText := http.StatusText(statusCode)
		if c.Debug {
			statusColor := colorForStatus(statusCode)
			Logger.Global.Printf("Error %s%d %s%s for %s \"%s\" from %s at %v (in %v)\n",
				statusColor, statusCode, statusText, AnsiReset, method, path, clientIp, end, latency)
		} else {
			Logger.Global.Printf("Error %d %s for %s \"%s\" from %s at %v (in %v)\n",
				statusCode, statusText, method, path, clientIp, end, latency)
		}
	}
}

// LogRequest logs a multiline message with information about each received request.
// One log line is emitted immediately when the request is received (in case of server crash),
// the remaining log lines are aggregated in a buffer allocated from a pool and only emitted
// after the request has been processed.
//
// Another middleware, like LogHeaders can access the request-specific logger from
// the *httpserver.Context with `c.GetLocal("Log")` or can use the `Logf("fmt", ..args)`,
// `LogValue(c, "Key", value)`, and `LogResponse(c, "Status", value)` helpers.
func LogRequest(c *httpserver.Context) {
	start := time.Now()
	path := c.Request.URL.Path
	method := c.Request.Method
	clientIp := c.ClientIp()

	// Always immediately log that we received a request, in case the request takes a long time
	Logger.Global.Printf("Received %s \"%s\" from %s at %v\n", method, path, clientIp, start.Format(LogTimeFormat))

	// Log preamble
	request := Logger.Pool.Get().(*RequestLog)
	request.Buffer.Reset()
	request.Printf("Log for %s \"%s\" from %s at %v\n", method, path, clientIp, start.Format(LogTimeFormat))
	defer Logger.Pool.Put(request)

	c.SetLocal("Log", request)
	defer delete(c.Locals, "Log")

	// Handle request
	c.PerformRequest()

	// Log postamble
	end := time.Now()
	latency := end.Sub(start)
	statusCode := c.Response.Status()
	statusText := http.StatusText(statusCode)
	if c.Debug {
		statusColor := colorForStatus(statusCode)
		request.Printf("Replied with %s%d %s%s in %v\n", statusColor, statusCode, statusText, AnsiReset, latency)
	} else {
		request.Printf("Replied with %d %s in %v\n", statusCode, statusText, latency)
	}

	// Write log
	Logger.Global.Print(request.Buffer.String())
}

// LogHeaders returns a middleware which logs any header values for headers in headerKeys.
// Header values longer than 60 characters are truncated with an ellipsis in the log output.
//
// ex. LogHeaders("Accept", "User-Agent")
func LogHeaders(headerKeys ...string) func(c *httpserver.Context) {
	return func(c *httpserver.Context) {
		for _, header := range headerKeys {
			value := c.Request.Header.Get(header)
			if len(value) > 0 {
				if len(value) <= 60 {
					Logger.LogValue(c, header, value)
				} else {
					Logger.LogValue(c, header, value[:56]+" ...")
				}
			}
		}

		c.ContinueRequest()
	}
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
