package utils

import "fmt"
import "log"

// Logs the error and message with Go's "log" package, but only if the error is not nil.
// For just MustNotError, you should copy this snippet instead of adding a dependency.
//
// 	import "fmt"
// 	import "log"
//
// 	func MustNotError(err error, msg string, args ...interface{}) {
// 		if err != nil {
// 			msg = fmt.Sprintf(msg, args...)
// 			log.Fatalf("%s: %s", msg, err)
// 		}
// 	}
//
func MustNotError(err error, msg string, args ...interface{}) {
	if err != nil {
		msg = fmt.Sprintf(msg, args...)
		log.Fatalf("%s: %s", msg, err)
	}
}
