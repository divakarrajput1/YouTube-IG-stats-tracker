package repository

import (
	"database/sql"
	"time"
)

// Video represents a video being tracked
type Video struct {
    ID        string    `db:"id" json:"id"`
    Platform  string    `db:"platform" json:"platform"`
    VideoID   string    `db:"video_id" json:"video_id"`
    InstagramUsername string `db:"instagram_username" json:"instagram_username,omitempty"`
    State     string    `db:"state" json:"state"`
    CreatedAt time.Time `db:"created_at" json:"created_at"`
    Tag       string    `db:"tag" json:"tag,omitempty"`
}

// VideoStats represents hourly stats for a video
type VideoStats struct {
    ID        string    `db:"id" json:"id"`
    VideoID   string    `db:"video_id" json:"video_id"`
    Timestamp time.Time `db:"timestamp" json:"timestamp"`
    Views     int       `db:"views" json:"views"`
    Likes     int       `db:"likes" json:"likes"`
    Comments    int       `db:"comments" json:"comments"`
    Caption     sql.NullString    `db:"caption" json:"caption,omitempty"`
    Permalink   sql.NullString    `db:"permalink" json:"permalink,omitempty"`
    MediaType   sql.NullString    `db:"media_type" json:"media_type,omitempty"`
    MediaURL    sql.NullString    `db:"media_url" json:"media_url,omitempty"`
}

// Platform types
const (
    PlatformYouTube   = "youtube"
    PlatformInstagram = "instagram"
)

// Video states
const (
    StateRegistered = "registered"
    StateArchived   = "archived"
    StateError      = "error"
)