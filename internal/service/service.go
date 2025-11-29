package service

import (
    "context"
    "fmt"
    "time"
    "video-stats-tracker/internal/platform"
    "video-stats-tracker/internal/repository"
)

type Service interface {
    // Core APIs
    RegisterVideo(ctx context.Context, platform, videoID, username, tag string) error
    GetVideoStats(ctx context.Context, videoID string, from, to time.Time) ([]repository.VideoStats, error)
    
    // Internal methods for polling
    GetAllVideos(ctx context.Context) ([]repository.Video, error)
    UpdateVideoStats(ctx context.Context, video *repository.Video) error
}

type videoService struct {
    repo            repository.Repository
    youtubeClient   *platform.YouTubeClient
    instagramClient *platform.InstagramClient
}

func NewService(repo repository.Repository, youtubeAPIKey, instagramToken, instagramID string) Service {
    return &videoService{
        repo:            repo,
        youtubeClient:   platform.NewYouTubeClient(youtubeAPIKey),
        instagramClient: platform.NewInstagramClient(instagramToken, instagramID),
    }
}

func (s *videoService) RegisterVideo(ctx context.Context, platform, videoID, username, tag string) error {
var existing *repository.Video
    var err error
    
    if platform == repository.PlatformInstagram {
        // For Instagram, check with username
        existing, err = s.repo.GetVideoWithUsername(ctx, platform, videoID, username)
    } else {
        // For YouTube, check normally
        existing, err = s.repo.GetVideo(ctx, platform, videoID)
    }
    
    if err != nil {
        return err
    }
    if existing != nil {
        return fmt.Errorf("video already being tracked")
    }

    // Validate platform
    if platform != repository.PlatformYouTube && platform != repository.PlatformInstagram {
        return fmt.Errorf("unsupported platform: %s", platform)
    }

    // For Instagram, validate that the post exists and get username
    if platform == repository.PlatformInstagram {
        if username == "" {
            return fmt.Errorf("instagram_username is required for Instagram posts")
        }
        _, err := s.instagramClient.GetMediaDetails(ctx, username, videoID)
        if err != nil {
            return fmt.Errorf("instagram post not found: %s for user %s", videoID, username)
        }
    }

    // Create new video
    video := &repository.Video{
        Platform: platform,
        VideoID:  videoID,
        InstagramUsername: username,
        State:    repository.StateRegistered,
        Tag:      tag,
    }

    return s.repo.CreateVideo(ctx, video)
}

func (s *videoService) GetVideoStats(ctx context.Context, videoID string, from, to time.Time) ([]repository.VideoStats, error) {
    return s.repo.GetVideoStats(ctx, videoID, from, to)
}

func (s *videoService) GetAllVideos(ctx context.Context) ([]repository.Video, error) {
    return s.repo.GetAllVideos(ctx)
}

func (s *videoService) UpdateVideoStats(ctx context.Context, video *repository.Video) error {
    var stats *repository.VideoStats
    var err error

    // Fetch stats from appropriate platform
    switch video.Platform {
    case repository.PlatformYouTube:
        stats, err = s.youtubeClient.GetVideoStats(ctx, video.VideoID)
    case repository.PlatformInstagram:
        stats, err = s.instagramClient.GetVideoStats(ctx, video.InstagramUsername, video.VideoID)
    default:
        return fmt.Errorf("unsupported platform: %s", video.Platform)
    }

    if err != nil {
        // Mark video as error if it's not found/accessible
        if err.Error() == "video not found" || err.Error() == "instagram post not found or inaccessible" {
            s.repo.UpdateVideoState(ctx, video.VideoID, repository.StateError)
        }
        return err
    }

    // Add timestamp and save
    stats.Timestamp = time.Now()
    return s.repo.CreateVideoStats(ctx, stats)
}