package main

import "net/http"

// Logging is a middleware stub; the middleware ticket fills it in.
func Logging(next http.Handler) http.Handler {
	return next
}

// Recover is a middleware stub; the middleware ticket fills it in.
func Recover(next http.Handler) http.Handler {
	return next
}

// BodyLimit is a middleware stub; the middleware ticket fills it in.
func BodyLimit(next http.Handler) http.Handler {
	return next
}

// ContentType is a middleware stub; the middleware ticket fills it in.
func ContentType(next http.Handler) http.Handler {
	return next
}
