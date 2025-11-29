package platform

import (
	"context"
    "database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"video-stats-tracker/internal/repository"
)

type InstagramClient struct {
    accessToken string
    client      *http.Client
    instagramID string // Instagram Business Account ID
}

func NewInstagramClient(accessToken, instagramID string) *InstagramClient {
    return &InstagramClient{
        accessToken: accessToken,
        client:      &http.Client{},
        instagramID: instagramID,
    }
}

// type InstagramResponse struct {
//     ID       string `json:"id"`
//     LikeCount int   `json:"like_count"`
// }

// func (i *InstagramClient) GetVideoStats(ctx context.Context, videoID string) (*repository.VideoStats, error) {
//     // Note: Instagram Graph API requires special permissions and app review
//     // This is a simplified implementation
//     url := fmt.Sprintf(
//         "https://graph.instagram.com/%s?fields=like_count&access_token=%s",
//         videoID, i.accessToken,
//     )

//     req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
//     if err != nil {
//         return nil, err
//     }

//     resp, err := i.client.Do(req)
//     if err != nil {
//         return nil, err
//     }
//     defer resp.Body.Close()

//     if resp.StatusCode != http.StatusOK {
//         return nil, fmt.Errorf("instagram API error: %s", resp.Status)
//     }

//     var igResp InstagramResponse
//     if err := json.NewDecoder(resp.Body).Decode(&igResp); err != nil {
//         return nil, err
//     }

//     // Instagram doesn't provide view count for regular posts
//     // For Reels/IGTV, you'd need additional API endpoints
//     return &repository.VideoStats{
//         VideoID: videoID,
//         Views:   0, // Instagram doesn't expose view count via basic API
//         Likes:   igResp.LikeCount,
//     }, nil
// }

// InstagramMedia represents the media data from business_discovery
type InstagramMedia struct {
    ID            string `json:"id"`
    LikeCount     int    `json:"like_count"`
    CommentsCount int    `json:"comments_count"`
    Caption       string `json:"caption"`
    MediaURL      string `json:"media_url"`
    Timestamp     string `json:"timestamp"`
    Permalink     string `json:"permalink"`
    MediaType     string `json:"media_type"`
}
// BusinessDiscoveryResponse represents the full API response
type BusinessDiscoveryResponse struct {
    BusinessDiscovery struct {
        Media struct {
            Data []InstagramMedia `json:"data"`
        } `json:"media"`
    } `json:"business_discovery"`
}

func (i *InstagramClient) GetVideoStats(ctx context.Context, username, videoID string) (*repository.VideoStats, error) {
    // Get ALL media from the configured Instagram account
    allMedia, err := i.getAllUserMedia(ctx, username)
    if err != nil {
        return nil, fmt.Errorf("failed to get user media: %v", err)
    }

    // Find the specific media we're looking for
    for _, media := range allMedia {
        if media.ID == videoID {
            fmt.Printf("✅ Found Instagram post - User: %s, Post: %s, Likes: %d, Comments: %d\n", 
                username,media.ID, media.LikeCount, media.CommentsCount)

            // For Instagram, we can use various metrics as views proxy
            // Using likes + comments as engagement metric for views
            engagement := media.LikeCount + media.CommentsCount
            if engagement == 0 {
                engagement = 1 // Minimum engagement to show activity
            }

            return &repository.VideoStats{
                VideoID: videoID,
                Views:   engagement, // Using engagement as proxy for views
                Likes:   media.LikeCount,
                Comments:           media.CommentsCount,
                Caption:            toNullString(media.Caption),
                Permalink:          toNullString(media.Permalink),
                MediaType:          toNullString(media.MediaType),
                MediaURL:           toNullString(media.MediaURL),
            }, nil
        }
    }

    return nil, fmt.Errorf("instagram post not found: %s for user %s", videoID, username)
}

// Helper function to convert string to sql.NullString
func toNullString(s string) sql.NullString {
    if s == "" {
        return sql.NullString{Valid: false}
    }
    return sql.NullString{String: s, Valid: true}
}
// getAllUserMedia fetches all media for the configured user
func (i *InstagramClient) getAllUserMedia(ctx context.Context, username string) ([]InstagramMedia, error) {
    url := fmt.Sprintf(
        "https://graph.facebook.com/v23.0/%s?fields=business_discovery.username(%s){media{id,like_count,comments_count,caption,media_url,timestamp,permalink,media_type}}&access_token=%s",
        i.instagramID, username, i.accessToken,
    )

    fmt.Printf("📡 Calling Instagram Business Discovery API for user: %s\n", username)

    req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
    if err != nil {
        return nil, err
    }

    resp, err := i.client.Do(req)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()

    body, _ := io.ReadAll(resp.Body)
    
    if resp.StatusCode != http.StatusOK {
        return nil, fmt.Errorf("instagram API error %d: %s", resp.StatusCode, string(body))
    }

    var discoveryResp BusinessDiscoveryResponse
    if err := json.Unmarshal(body, &discoveryResp); err != nil {
        return nil, fmt.Errorf("failed to parse Instagram response: %v", err)
    }

    fmt.Printf("✅ Found %d media posts for Instagram user: %s\n", 
        len(discoveryResp.BusinessDiscovery.Media.Data), username)

    return discoveryResp.BusinessDiscovery.Media.Data, nil
}

// GetMediaDetails returns full media details for a specific post
func (i *InstagramClient) GetMediaDetails(ctx context.Context, username,videoID string) (*InstagramMedia, error) {
    allMedia, err := i.getAllUserMedia(ctx, username)
    if err != nil {
        return nil, err
    }

    for _, media := range allMedia {
        if media.ID == videoID {
            return &media, nil
        }
    }

    return nil, fmt.Errorf("media not found: %s for user %s", videoID, username)
}