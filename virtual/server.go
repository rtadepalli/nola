package virtual

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/richardartoul/nola/virtual/registry"
)

type server struct {
	// Dependencies.
	registry    registry.Registry
	environment Environment
}

// NewServer creates a new server for the actor virtual environment.
func NewServer(
	registry registry.Registry,
	environment Environment,
) *server {
	return &server{
		registry:    registry,
		environment: environment,
	}
}

// Start starts the server.
func (s *server) Start(port int) error {
	http.HandleFunc("/api/v1/register-module", s.registerModule)
	http.HandleFunc("/api/v1/create-actor", s.createActor)
	http.HandleFunc("/api/v1/invoke", s.invoke)

	if err := http.ListenAndServe(fmt.Sprintf(":%d", port), nil); err != nil {
		return err
	}

	return nil
}

// This one is a bit weird because its basically a file upload with some JSON
// so we just shove the JSON into the headers cause I'm lazy to do anything
// more clever.
func (s *server) registerModule(w http.ResponseWriter, r *http.Request) {
	var (
		namespace = r.Header.Get("namespace")
		moduleID  = r.Header.Get("module_id")
	)

	moduleBytes, err := ioutil.ReadAll(io.LimitReader(r.Body, 1<<24))
	if err != nil {
		w.WriteHeader(500)
		w.Write([]byte(err.Error()))
		return
	}

	ctx, cc := context.WithTimeout(context.Background(), 60*time.Second)
	defer cc()
	result, err := s.registry.RegisterModule(ctx, namespace, moduleID, moduleBytes, registry.ModuleOptions{})
	if err != nil {
		w.WriteHeader(500)
		w.Write([]byte(err.Error()))
		return
	}

	marshaled, err := json.Marshal(result)
	if err != nil {
		w.WriteHeader(500)
		w.Write([]byte(err.Error()))
		return
	}

	w.WriteHeader(200)
	w.Write(marshaled)
}

type createActorRequest struct {
	Namespace string `json:"namespace"`
	ActorID   string `json:"actor_id"`
	ModuleID  string `json:"module_id"`
}

func (s *server) createActor(w http.ResponseWriter, r *http.Request) {
	jsonBytes, err := ioutil.ReadAll(io.LimitReader(r.Body, 1<<20))
	if err != nil {
		w.WriteHeader(500)
		w.Write([]byte(err.Error()))
		return
	}

	var req createActorRequest
	if err := json.Unmarshal(jsonBytes, &req); err != nil {
		w.WriteHeader(500)
		w.Write([]byte(err.Error()))
		return
	}

	ctx, cc := context.WithTimeout(context.Background(), 5*time.Second)
	defer cc()
	result, err := s.registry.CreateActor(ctx, req.Namespace, req.ActorID, req.ModuleID, registry.ActorOptions{})
	if err != nil {
		w.WriteHeader(500)
		w.Write([]byte(err.Error()))
		return
	}

	marshaled, err := json.Marshal(result)
	if err != nil {
		w.WriteHeader(500)
		w.Write([]byte(err.Error()))
		return
	}

	w.WriteHeader(200)
	w.Write(marshaled)
}

type invokeRequest struct {
	Namespace string `json:"namespace"`
	ActorID   string `json:"actor_id"`
	Operation string `json:"operation"`
	Payload   []byte `json:"payload"`
}

func (s *server) invoke(w http.ResponseWriter, r *http.Request) {
	jsonBytes, err := ioutil.ReadAll(io.LimitReader(r.Body, 1<<24))
	if err != nil {
		w.WriteHeader(500)
		w.Write([]byte(err.Error()))
		return
	}

	var req invokeRequest
	if err := json.Unmarshal(jsonBytes, &req); err != nil {
		w.WriteHeader(500)
		w.Write([]byte(err.Error()))
		return
	}

	// TODO: This should be configurable, probably in a header with some maximum.
	ctx, cc := context.WithTimeout(context.Background(), 5*time.Second)
	defer cc()
	result, err := s.environment.Invoke(ctx, req.Namespace, req.ActorID, req.Operation, req.Payload)
	if err != nil {
		w.WriteHeader(500)
		w.Write([]byte(err.Error()))
		return
	}

	w.WriteHeader(200)
	w.Write(result)
}