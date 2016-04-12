// Package vanilla is Reflexion Health's Golang standard library.
// Many of the libraries are focused on supporting web services by providing
// additional support for sql, http, json, etc...
//
// Packages:
//   crypto - wrapper around multiple crypto/etc packages for common crypto operations
//   date - an extension to the built-in time library to deal explicitly with dates
//   httpserver - a fork of httpserver's server / gin-gonic's engine that works the way we like
//   httpserver/stack - utilities and middleware for the reflexion http server
//   null - nullable types that support database/sql, encoding/gob, and encoding/json
//   semver - helper library for working with semantic versioning
//   sql - utilities to make working with sql in Golang more natural
//   sql/language - ast, parser, et. al. for handling multiple sql dialects
package vanilla
