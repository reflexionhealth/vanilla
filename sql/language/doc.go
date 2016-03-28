/*
Package language contains an ast and parser for the sql language.

It differs from other parsers by having a configruable dialect, so that it can
be used to validate, understand, and test queries for multiple other databases.

For a "complete", fast, general-purpose parser, consider adapting code from
an actual database like [cockroachdb](https://github.com/cockroachdb/cockroach).

Currently supported behavior:
 + Parsing simple SELECT statements with expressions
 + Expressions have correct operator precedence in each dialect
 + Syntax validation (but not semantic validation)

*/
package language
