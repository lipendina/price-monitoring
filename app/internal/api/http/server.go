package http

import (
	"avito/internal/service"
	"context"
	"fmt"
	"github.com/gorilla/mux"
	"log"
	"net/http"
	"os"
	"time"
)

type Server struct {
	logger *log.Logger
	srv    *http.Server
	port   uint16
}

func (s *Server) Start() {
	s.logger.Printf("Server started on port %d\n", s.port)
	go func() {
		if err := s.srv.ListenAndServe(); err != nil {
			if err.Error() != http.ErrServerClosed.Error() {
				s.logger.Fatalf("Start http server error, reason: %+v\n", err)
			}
		}
	}()
}

func (s *Server) Stop() {
	s.logger.Printf("Server stopped")
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	if err := s.srv.Shutdown(ctx); err != nil {
		s.logger.Printf("Http server shutdown error, reason: %+v\n", err)
	}
}

func NewServer(port uint16, service service.ServiceAPI) *Server {
	logger := log.New(os.Stdout, "HTTP: ", log.Ldate|log.Ltime|log.Lshortfile)
	handler := newHandler(service)

	return &Server{
		logger: logger,
		port:   port,
		srv: &http.Server{
			Addr:         fmt.Sprintf(":%d", port),
			Handler:      handler,
			ReadTimeout:  10 * time.Second,
			WriteTimeout: 10 * time.Second,
		},
	}
}

func newHandler(service service.ServiceAPI) http.Handler {
	logger := log.New(os.Stdout, "CONTROLLER: ", log.Ldate|log.Ltime|log.Lshortfile)
	controller := NewController(service, logger)

	router := mux.NewRouter()

	router.Use(PanicRecovery)

	router.HandleFunc("/subscribe", controller.PriceChangeSubscriptionHandler).Methods("POST")
	router.HandleFunc("/confirm", controller.ConfirmationHandler).Methods("GET")
	router.HandleFunc("/unsubscribe", controller.UnsubscriptionHandler).Methods("POST")

	return router
}
