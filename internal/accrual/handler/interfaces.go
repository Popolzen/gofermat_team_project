package handler

import "net/http"

type AccrualInterface interface {
	GetOrder(w http.ResponseWriter, r *http.Request)
	RegisterOrder(w http.ResponseWriter, r *http.Request)
}
