package server

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
)

// httpServer wraps the log and handles HTTP requests
type httpServer struct {
	Log *Log
}

// newHTTPServer creates a new HTTP server instance with an initialized log
func newHTTPServer() *httpServer {
	return &httpServer{
		Log: NewLog(),
	}
}

// ---- Produce/Consume Request/Response ----

type ProduceRequest struct {
	Record Record `json:"record"`
}

type ProduceResponse struct {
	Offset uint64 `json:"offset"`
}

type ConsumeRequest struct {
	Offset uint64 `json:"offset"`
}

type ConsumeResponse struct {
	Record Record `json:"record"`
}

// NewHTTPServer creates and configures a new HTTP server with routing
func NewHTTPServer(addr string) *http.Server {
	httpsrv := newHTTPServer()
	r := mux.NewRouter()

	// Set up routes: GET for consuming, POST for producing
	r.HandleFunc("/", httpsrv.handleConsume).Methods("GET")
	r.HandleFunc("/", httpsrv.handleProduce).Methods("POST")

	return &http.Server{
		Addr:    addr,
		Handler: r,
	}
}

// handleProduce processes POST requests to add new records to the log
func (s *httpServer) handleProduce(w http.ResponseWriter, r *http.Request) {
	var req ProduceRequest

	// Parse JSON request body
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Add record to log and get its offset
	off, err := s.Log.Append(req.Record)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Return the offset of the newly added record
	res := ProduceResponse{
		Offset: off,
	}

	if err := json.NewEncoder(w).Encode(res); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// handleConsume processes GET requests to retrieve records from the log
func (s *httpServer) handleConsume(w http.ResponseWriter, r *http.Request) {
	var req ConsumeRequest

	// Parse JSON request body to get the desired offset
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Retrieve record at the specified offset
	record, err := s.Log.Read(req.Offset)
	if err == ErrOffsetNotFound {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Return the requested record
	res := ConsumeResponse{
		Record: record,
	}

	err = json.NewEncoder(w).Encode(res)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
