// Package vanilla is Reflexion Health's Golang standard library.
// Many of the libraries are focused on supporting web services by providing
// additional support for sql, http, json, etc...
//
// Packages:
//   crypto - wrapper around multiple crypto/etc packages for common crypto operations
//   date - an extension to the built-in time library to deal explicitly with dates
//   httpserver - a fork of httpserver's server / gin-gonic's engine that works the way we like
//   httpserver/stack - utilities and middleware for the reflexion http server
//   math - (WIP) helper library for math operations
//   null - nullable types that support database/sql, encoding/gob, and encoding/json
//   semver - (WIP) helper library for working with semantic versioning
//   sql - utilities to make working with sql in Golang more natural
package vanilla

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
