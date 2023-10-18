package di_test

import (
	"io"
	"log"
	"net/http"

	"github.com/mcvoid/di"
)

// Example interface that abstracts a logger.
type Logger interface {
	Fatal(v ...any)
}

// EchoHandler is an HTTP Handler that echoes its input.
type EchoHandler struct{}

// ServeHTTP handles an HTTP request to the server.
func (*EchoHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	io.Copy(w, r.Body)
}

// server is an injectable server struct
type server struct {
	logger      Logger
	echoHandler *EchoHandler
}

// Starts the server
func (s *server) Serve() {
	s.logger.Fatal(http.ListenAndServe(":8080", s.echoHandler))
}

// This is what Inject calls to inject the dependencies
func (s *server) Bind(logger Logger, echo *EchoHandler) {
	s.logger = logger
	s.echoHandler = echo
}

// Demonstrates using Context to run a function with injectable arguments.
func ExampleContext() {
	// First create the context and add your dependencies.
	ctx := di.New().Add(log.Default()).Add(&EchoHandler{})

	// Method 1: Use a function and have the dependencies be
	// injected through its parameters.
	startServer := func(logger Logger, echo *EchoHandler) {
		logger.Fatal(http.ListenAndServe(":8080", echo))
	}
	if err := ctx.Inject(startServer); err != nil {
		log.Fatal(err)
	}

	// Method 2: Use an object with the Bind method.
	s := &server{}
	if err := ctx.Inject(s); err != nil {
		log.Fatal(err)
	}
	s.Serve()
}
