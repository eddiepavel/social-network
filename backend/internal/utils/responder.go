package utils

import (
	"encoding/json"
	"net/http"
)

type Envelope[T any] struct {
	Data       *T          `json:"data,omitempty"`
	Error      *ErrBody    `json:"error,omitempty"`
	Pagination *Pagination `json:"pagination,omitempty"`
}

type ErrBody struct {
	Code    string      `json:"code"`
	Message string      `json:"message"`
	Details interface{} `json:"details,omitempty"`
}

type Pagination struct {
	Page       int `json:"page"`
	Size       int `json:"size"`
	Current    int `json:"current"`
	TotalItems int `json:"total_items"`
	TotalPages int `json:"total_pages"`
}

func Write(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func OK[T any](w http.ResponseWriter, data T) {
	Write(w, http.StatusOK, Envelope[T]{Data: &data})
}

func Created[T any](w http.ResponseWriter, location string, data T) {
	Write(w, http.StatusCreated, Envelope[T]{Data: &data})
}

func NoContent(w http.ResponseWriter) {
	w.WriteHeader(http.StatusNoContent)
}

func WithPagination[T any](data T, p Pagination) Envelope[T] {
	return Envelope[T]{Data: &data, Pagination: &p}
}

func Error(w http.ResponseWriter, status int, code, msg string, details any) {
	Write(w, status, Envelope[struct{}]{Error: &ErrBody{
		Code: code, Message: msg, Details: details,
	}})
}

func BadRequest(w http.ResponseWriter, err error) {
	Error(w, http.StatusBadRequest, "400", "Invalid request.", errText(err))
}

func Unauthorized(w http.ResponseWriter, reason string) {
	Error(w, http.StatusUnauthorized, "401", "Authentication required.", reason)
}

func Forbidden(w http.ResponseWriter) {
	Error(w, http.StatusForbidden, "403", "You don't have access to this resource.", nil)
}

func NotFound(w http.ResponseWriter) {
	Error(w, http.StatusNotFound, "404", "Resource not found.", nil)
}

func Internal(w http.ResponseWriter, err error) {
	Error(w, http.StatusInternalServerError, "500", "Something went wrong.", errText(err))
}

func errText(err error) string {
	if err == nil {
		return ""
	}
	return err.Error()
}
