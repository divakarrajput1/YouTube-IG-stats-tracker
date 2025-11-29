package endpoint

import (
    "context"
    "time"

    "github.com/go-kit/kit/endpoint"
    "video-stats-tracker/internal/service"
)

type Endpoints struct {
    RegisterVideo endpoint.Endpoint
    GetStats      endpoint.Endpoint
}

type RegisterVideoRequest struct {
    Platform string `json:"platform"`
    VideoID  string `json:"video_id"`
    Username  string `json:"username,omitempty"`
    Tag      string `json:"tag,omitempty"`
}

type RegisterVideoResponse struct {
    ID    string `json:"id,omitempty"`
    Error string `json:"error,omitempty"`
}

type GetStatsRequest struct {
    VideoID string    `json:"video_id"`
    From    time.Time `json:"from"`
    To      time.Time `json:"to"`
}

type GetStatsResponse struct {
    Stats []interface{} `json:"stats"` // Using interface{} for flexibility
    Error string        `json:"error,omitempty"`
}

func MakeEndpoints(s service.Service) Endpoints {
    return Endpoints{
        RegisterVideo: makeRegisterVideoEndpoint(s),
        GetStats:      makeGetStatsEndpoint(s),
    }
}

func makeRegisterVideoEndpoint(s service.Service) endpoint.Endpoint {
    return func(ctx context.Context, request interface{}) (interface{}, error) {
        req := request.(RegisterVideoRequest)
        err := s.RegisterVideo(ctx, req.Platform, req.VideoID, req.Username, req.Tag)
        if err != nil {
            return RegisterVideoResponse{Error: err.Error()}, nil
        }
        return RegisterVideoResponse{}, nil
    }
}

func makeGetStatsEndpoint(s service.Service) endpoint.Endpoint {
    return func(ctx context.Context, request interface{}) (interface{}, error) {
        req := request.(GetStatsRequest)
        stats, err := s.GetVideoStats(ctx, req.VideoID, req.From, req.To)
        if err != nil {
            return GetStatsResponse{Error: err.Error()}, nil
        }
        
        // Convert to interface slice for JSON marshaling
        interfaceStats := make([]interface{}, len(stats))
        for i, stat := range stats {
            interfaceStats[i] = stat
        }
        
        return GetStatsResponse{Stats: interfaceStats}, nil
    }
}