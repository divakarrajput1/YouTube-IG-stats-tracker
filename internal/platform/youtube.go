package platform

import (
    "context"
    "encoding/json"
    "fmt"
    "net/http"
    "strconv"
    "video-stats-tracker/internal/repository"
)

type YouTubeClient struct {
    apiKey string
    client *http.Client
}

func NewYouTubeClient(apiKey string) *YouTubeClient {
    return &YouTubeClient{
        apiKey: apiKey,
        client: &http.Client{},
    }
}

type YouTubeResponse struct {
    Items []struct {
        Statistics struct {
            ViewCount string `json:"viewCount"`
            LikeCount string `json:"likeCount"`
        } `json:"statistics"`
    } `json:"items"`
}

func (y *YouTubeClient) GetVideoStats(ctx context.Context, videoID string) (*repository.VideoStats, error) {
    url := fmt.Sprintf(
        "https://www.googleapis.com/youtube/v3/videos?part=statistics&id=%s&key=%s",
        videoID, y.apiKey,
    )

    req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
    if err != nil {
        return nil, err
    }

    resp, err := y.client.Do(req)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()

    var ytResp YouTubeResponse
    if err := json.NewDecoder(resp.Body).Decode(&ytResp); err != nil {
        return nil, err
    }

    if len(ytResp.Items) == 0 {
        return nil, fmt.Errorf("video not found: %s", videoID)
    }

    stats := ytResp.Items[0].Statistics
    
    // Convert string counts to integers
    views, _ := strconv.Atoi(stats.ViewCount)
    likes, _ := strconv.Atoi(stats.LikeCount)

    return &repository.VideoStats{
        VideoID: videoID,
        Views:   views,
        Likes:   likes,
    }, nil
}