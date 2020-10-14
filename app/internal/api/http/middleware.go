package http

import (
	"log"
	"net/http"
)

func PanicRecovery(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				log.Println(err)
				w.WriteHeader(http.StatusInternalServerError)
			}
		}()
		handler.ServeHTTP(w, r)
	})
}
