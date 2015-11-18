// This file is Copyright 2014 Manu Martinez-Almeida.  All rights reserved.
// Use of this source code is governed by a MIT style license.
//
// Modifications by Kevin Stenerson for Reflexion Health Inc. Copyright 2015
package httpbase

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"regexp"
	"runtime"

	"github.com/reflexionhealth/vanilla/router"
)

var (
	// stackFilters is a list of Regexps to filter out lines in the callstack
	stackFilters = []*regexp.Regexp{
		regexp.MustCompile(`internal/router/context`),
		regexp.MustCompile(`internal/router/router`),
		regexp.MustCompile(`net/http/server`),
		regexp.MustCompile(`go/src/runtime`),
	}
)

var (
	dunno     = []byte("???")
	centerDot = []byte("·")
	dot       = []byte(".")
	slash     = []byte("/")
)

// Recover is a middlerware that recovers from any panics and writes a 500 if there was one.
// Logs to the specified writter buffer. If nil is provided, it will still recover, but won't log.
// Example: os.Stdout, a file opened in write mode, a socket...
func Recover(c *router.Context) {
	// Use "defer" so we can capture a panic
	defer func() {
		if err := recover(); err != nil {
			stack := stack(4)
			Logger.LogResponse(c, "Panic", err)
			Logger.Logf(c, "%s\n", stack)

			if !c.Response.Rendered() {
				c.Response.Header().Set("Content-Type", "application/json; charset=utf-8")
				c.Response.Header().Set("Reflexion-Request-Errors", "[\"Something went wrong\"]")
				c.Response.JSON(500, "{\"errors\":[\"Something went wrong\"]}")
			} else {
				Logger.Logf(c, "\n  Panic occured after write: error not included in response\n")
			}
		}
	}()

	// Call the next handler
	c.MustContinue() // only use MustContinue for performance critical middleware
}

// stack returns a nicely formated stack frame, skipping "skip" frames
func stack(skip int) []byte {
	buf := new(bytes.Buffer) // the returned data
	// As we loop, we open files and read them. These variables record the currently
	// loaded file.
	var lines [][]byte
	var lastFile string
	for i := skip; ; i++ { // Skip the expected number of frames
		pc, file, line, ok := runtime.Caller(i)
		if !ok {
			break
		}

		shouldSkip := false
		for _, exp := range stackFilters {
			if exp.MatchString(file) {
				shouldSkip = true
			}
		}

		// Skip filtered lines, but only if they aren't in the top two lines of the callstack
		if shouldSkip && i > (skip+1) {
			continue
		}

		// Print this much at least.  If we can't find the source, it won't show.
		fmt.Fprintf(buf, "\n  %s:%d (0x%x)\n", file, line, pc)
		if file != lastFile {
			data, err := ioutil.ReadFile(file)
			if err != nil {
				continue
			}
			lines = bytes.Split(data, []byte{'\n'})
			lastFile = file
		}
		fmt.Fprintf(buf, "\t%s: %s\n", function(pc), source(lines, line))
	}
	return buf.Bytes()
}

// source returns a space-trimmed slice of the n'th line.
func source(lines [][]byte, n int) []byte {
	n-- // in stack trace, lines are 1-indexed but our array is 0-indexed
	if n < 0 || n >= len(lines) {
		return dunno
	}
	return bytes.TrimSpace(lines[n])
}

// function returns, if possible, the name of the function containing the PC.
func function(pc uintptr) []byte {
	fn := runtime.FuncForPC(pc)
	if fn == nil {
		return dunno
	}
	name := []byte(fn.Name())

	// The name includes the path name to the package, which is unnecessary
	// since the file name is already included.  Plus, it has center dots.
	//
	// That is, we see:
	//
	//	runtime/debug.*T·ptrmethod
	//
	// and want:
	//
	//	*T.ptrmethod
	//
	// Also the package path might contains dot (e.g. code.google.com/...),
	// so first eliminate the path prefix
	if lastslash := bytes.LastIndex(name, slash); lastslash >= 0 {
		name = name[lastslash+1:]
	}
	if period := bytes.Index(name, dot); period >= 0 {
		name = name[period+1:]
	}
	name = bytes.Replace(name, centerDot, dot, -1)
	return name
}
