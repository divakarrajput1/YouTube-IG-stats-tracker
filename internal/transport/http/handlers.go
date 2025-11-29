package http

import (
    "context"
    "encoding/json"
    "net/http"
	"os"
    "time"

	"github.com/go-kit/log"
    "github.com/gorilla/mux"
    "github.com/go-kit/kit/transport"
    kitHttp "github.com/go-kit/kit/transport/http"

    "video-stats-tracker/internal/endpoint"
)

func NewHTTPHandler(endpoints endpoint.Endpoints) http.Handler {
    r := mux.NewRouter()
    logger := log.NewLogfmtLogger(os.Stderr)
    options := []kitHttp.ServerOption{
        kitHttp.ServerErrorHandler(transport.NewLogErrorHandler(logger)),
        kitHttp.ServerErrorEncoder(encodeError),
    }

    // Register video endpoint
    r.Methods("POST").Path("/track-video").Handler(kitHttp.NewServer(
        endpoints.RegisterVideo,
        decodeRegisterVideoRequest,
        encodeResponse,
        options...,
    ))

    // Get stats endpoint
    r.Methods("GET").Path("/stats").Handler(kitHttp.NewServer(
        endpoints.GetStats,
        decodeGetStatsRequest,
        encodeResponse,
        options...,
    ))

    return r
}

func decodeRegisterVideoRequest(_ context.Context, r *http.Request) (interface{}, error) {
    var req endpoint.RegisterVideoRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        return nil, err
    }
    return req, nil
}

func decodeGetStatsRequest(_ context.Context, r *http.Request) (interface{}, error) {
    var req endpoint.GetStatsRequest
    req.VideoID = r.URL.Query().Get("video_id")
    
    // Parse from date (default to 7 days ago)
    fromStr := r.URL.Query().Get("from")
    if fromStr == "" {
        req.From = time.Now().AddDate(0, 0, -7)
    } else {
        from, err := time.Parse(time.RFC3339, fromStr)
        if err != nil {
            return nil, err
        }
        req.From = from
    }
    
    // Parse to date (default to now)
    toStr := r.URL.Query().Get("to")
    if toStr == "" {
        req.To = time.Now()
    } else {
        to, err := time.Parse(time.RFC3339, toStr)
        if err != nil {
            return nil, err
        }
        req.To = to
    }
    
    return req, nil
}

func encodeResponse(ctx context.Context, w http.ResponseWriter, response interface{}) error {
    w.Header().Set("Content-Type", "application/json; charset=utf-8")
    return json.NewEncoder(w).Encode(response)
}

func encodeError(_ context.Context, err error, w http.ResponseWriter) {
    w.Header().Set("Content-Type", "application/json; charset=utf-8")
    w.WriteHeader(http.StatusInternalServerError)
    json.NewEncoder(w).Encode(map[string]interface{}{
        "error": err.Error(),
    })
}