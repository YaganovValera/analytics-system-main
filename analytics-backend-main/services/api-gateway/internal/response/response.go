// api-gateway/internal/response/response.go
package response

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
)

type errorBody struct {
	Error struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
	} `json:"error"`
}

// JSON пишет успешный JSON-ответ.
func JSON(w http.ResponseWriter, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(v)
}

// writeError пишет JSON-ошибку с заданным статусом и сообщением.
func writeError(w http.ResponseWriter, code int, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	res := errorBody{}
	res.Error.Code = code
	res.Error.Message = msg
	_ = json.NewEncoder(w).Encode(res)
}

// Стандартные ответы
func BadRequest(w http.ResponseWriter, msg string)   { writeError(w, http.StatusBadRequest, msg) }
func Unauthorized(w http.ResponseWriter, msg string) { writeError(w, http.StatusUnauthorized, msg) }
func InternalError(w http.ResponseWriter, msg string) {
	writeError(w, http.StatusInternalServerError, msg)
}

// DebugError пишет детали ошибки, если включен debug-режим
func DebugError(w http.ResponseWriter, code int, err error, context string) {
	debug := os.Getenv("DEBUG") == "true"
	if debug {
		msg := fmt.Sprintf("%s: %v", context, err)
		log.Printf("debug error: %v", err.Error())
		writeError(w, code, msg)
	} else {
		writeError(w, code, context)
	}
}

// SafeError пишет ошибку, но вытаскивает context или fallback message
func SafeError(w http.ResponseWriter, code int, err error, fallback string) {
	msg := fallback
	if err != nil {
		var e *json.UnmarshalTypeError
		if errors.As(err, &e) {
			msg = fmt.Sprintf("JSON type error: field '%s'", e.Field)
		} else if errors.Is(err, context.DeadlineExceeded) {
			msg = "request timeout"
		}
	}
	writeError(w, code, msg)
}
