// Package vanilla is Reflexion Health's Golang standard library.
// Many of the libraries are focused on supporting web services by providing
// additional support for sql, http, json, etc...
//
// Packages:
//   crypto - wrapper around multiple crypto/etc packages for common crypto operations
//   date - an extension to the built-in time library to deal explicitly with dates
//   http - utilities and middleware for the reflexion http router
//   math - (WIP) helper library for math operations
//   router - a fork of httprouter's router / gin-gonic's engine that works the way we like
//   semver - (WIP) helper library for working with semantic versioning
//   sql - utilities to make working with sql in Golang more natural
//   sql/nullable - nullable types that support both database/sql and encoding/json
package vanilla

import "fmt"
import "log"

// Logs the error and message with Go's "log" package, but only if the error is not nil.
//
// Typically, it is easier/cleaner to just copy the following than import this package
//
// 	import "log"
//
// 	func checkf(err error, msg string, args ...interface{}) {
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
