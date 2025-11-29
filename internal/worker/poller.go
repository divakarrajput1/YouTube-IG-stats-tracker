package worker

import (
	"context"
	"log"
	"time"

	"video-stats-tracker/internal/repository"
	"video-stats-tracker/internal/service"

	"github.com/robfig/cron/v3"
)

type Poller struct {
    service    service.Service
    cron       *cron.Cron
    viralVideos map[string]time.Time // Track viral videos and their detection time
}

func NewPoller(service service.Service) *Poller {
    return &Poller{
        service:     service,
        cron:        cron.New(),
        viralVideos: make(map[string]time.Time),
    }
}

func (p *Poller) Start() {
    // Regular polling every hour
    p.cron.AddFunc("@every 2m", p.pollAllVideos)
    
    // Viral detection polling every 5 minutes
    p.cron.AddFunc("@every 1m", p.pollViralVideos)
    
    p.cron.Start()
    log.Println("Polling worker started")
}

func (p *Poller) Stop() {
    p.cron.Stop()
    log.Println("Polling worker stopped")
}

func (p *Poller) pollAllVideos() {
    ctx := context.Background()
    
    videos, err := p.service.GetAllVideos(ctx)
    if err != nil {
        log.Printf("Error fetching videos: %v", err)
        return
    }

    for _, video := range videos {
        if err := p.service.UpdateVideoStats(ctx, &video); err != nil {
            log.Printf("Error updating stats for video %s: %v", video.VideoID, err)
            continue
        }
        
        // Check for viral condition (simplified)
        p.checkViralCondition(ctx, &video)
    }
}

func (p *Poller) pollViralVideos() {
   // ctx := context.Background()
    
    // Poll only viral videos more frequently
    for videoID := range p.viralVideos {
        // Clean up old viral videos (more than 24 hours)
        if time.Since(p.viralVideos[videoID]) > 24*time.Hour {
            delete(p.viralVideos, videoID)
            continue
        }
        
        // In real implementation, you'd fetch and update viral videos here
        log.Printf("Polling viral video: %s", videoID)
    }
}

func (p *Poller) checkViralCondition(ctx context.Context, video *repository.Video) {
    // Simplified viral detection - in real implementation, compare with previous stats
    log.Printf("Checking viral condition for video: %s", video.VideoID)
    
    // Example: Mark as viral (you'd implement proper spike detection here)
    p.viralVideos[video.VideoID] = time.Now()
}