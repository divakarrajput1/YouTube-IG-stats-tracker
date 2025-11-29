package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
	_ "modernc.org/sqlite" // 👈 Changed to pure Go SQLite
)

type Repository interface {
    // Video operations
    CreateVideo(ctx context.Context, video *Video) error
    GetVideo(ctx context.Context, platform, videoID string) (*Video, error)
    GetVideoWithUsername(ctx context.Context, platform, videoID, username string) (*Video, error)
    GetAllVideos(ctx context.Context) ([]Video, error)
    UpdateVideoState(ctx context.Context, videoID, state string) error
    
    // Stats operations
    CreateVideoStats(ctx context.Context, stats *VideoStats) error
    GetVideoStats(ctx context.Context, videoID string, from, to time.Time) ([]VideoStats, error)
    GetLatestStats(ctx context.Context, videoID string) (*VideoStats, error)
}

type sqliteRepository struct {
    db *sqlx.DB
}

// NewSQLiteRepository creates a new SQLite repository
func NewSQLiteRepository() (Repository, error) {
    db, err := sqlx.Open("sqlite", "./video_stats.db")  // 👈 Changed to "sqlite" (no 3)
    if err != nil {
        return nil, err
    }
    
    // Test connection
    if err := db.Ping(); err != nil {
        return nil, err
    }
    
    if err := createTables(db); err != nil {
        return nil, err
    }
    // Run migration to add new columns
    if err := migrateDatabase(db); err != nil {
        return nil, err
    }
    return &sqliteRepository{db: db}, nil
}

func createTables(db *sqlx.DB) error {
    videosTable := `
    CREATE TABLE IF NOT EXISTS videos (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        platform VARCHAR(20) NOT NULL,
        video_id VARCHAR(100) NOT NULL,
        instagram_username VARCHAR(100),
        state VARCHAR(20) DEFAULT 'registered',
        created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
        tag VARCHAR(100),
        UNIQUE(platform, video_id, instagram_username)
    )`

    statsTable := `
    CREATE TABLE IF NOT EXISTS video_stats (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        video_id VARCHAR(100) NOT NULL,
        timestamp DATETIME DEFAULT CURRENT_TIMESTAMP,
        views INTEGER NOT NULL DEFAULT 0,
        likes INTEGER NOT NULL DEFAULT 0,
        comments INTEGER NOT NULL DEFAULT 0,
        caption TEXT,
        permalink TEXT,
        media_type TEXT,
        media_url TEXT
    )`

    _, err := db.Exec(videosTable)
    if err != nil {
        return err
    }

    _, err = db.Exec(statsTable)
    return err
}

func (r *sqliteRepository) CreateVideo(ctx context.Context, video *Video) error {

    if video.Platform == PlatformInstagram && video.InstagramUsername == "" {
        return fmt.Errorf("instagram_username is required for Instagram videos")
    }
    query := `
    INSERT INTO videos (platform, video_id, instagram_username, state, tag) 
    VALUES (?, ?, ?, ?, ?)`
    
    result, err := r.db.ExecContext(ctx, query, video.Platform, video.VideoID, video.InstagramUsername, video.State, video.Tag)
    if err != nil {
        return err
    }
    
    id, err := result.LastInsertId()
    if err != nil {
        return err
    }
    
    video.ID = fmt.Sprintf("%d", id)
    video.CreatedAt = time.Now()
    return nil
}

func (r *sqliteRepository) GetVideo(ctx context.Context, platform, videoID string) (*Video, error) {
    var video Video
    query := `SELECT * FROM videos WHERE platform = ? AND video_id = ?`
    err := r.db.GetContext(ctx, &video, query, platform, videoID)
    if err == sql.ErrNoRows {
        return nil, nil
    }
    return &video, err
}

// Add this new method to get video with Instagram username
func (r *sqliteRepository) GetVideoWithUsername(ctx context.Context, platform, videoID, username string) (*Video, error) {
    var video Video
    query := `SELECT * FROM videos WHERE platform = ? AND video_id = ? AND instagram_username = ?`
    err := r.db.GetContext(ctx, &video, query, platform, videoID, username)
    if err == sql.ErrNoRows {
        return nil, nil
    }
    return &video, err
}
func (r *sqliteRepository) GetAllVideos(ctx context.Context) ([]Video, error) {
    var videos []Video
    query := `SELECT * FROM videos WHERE state = 'registered'`
    err := r.db.SelectContext(ctx, &videos, query)
    return videos, err
}

func (r *sqliteRepository) UpdateVideoState(ctx context.Context, videoID, state string) error {
    query := `UPDATE videos SET state = ? WHERE video_id = ?`
    _, err := r.db.ExecContext(ctx, query, state, videoID)
    return err
}

func (r *sqliteRepository) CreateVideoStats(ctx context.Context, stats *VideoStats) error {
    query := `
    INSERT INTO video_stats (video_id, timestamp, views, likes, comments, caption, permalink, media_type, media_url) 
    VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`

    result, err := r.db.ExecContext(ctx, query, stats.VideoID, stats.Timestamp, stats.Views, stats.Likes, stats.Comments, stats.Caption, stats.Permalink, stats.MediaType, stats.MediaURL)
    if err != nil {
        return err
    }
    
    id, err := result.LastInsertId()
    if err != nil {
        return err
    }
    
    stats.ID = fmt.Sprintf("%d", id)
    return nil
}

func (r *sqliteRepository) GetVideoStats(ctx context.Context, videoID string, from, to time.Time) ([]VideoStats, error) {
    var stats []VideoStats
    query := `SELECT * FROM video_stats WHERE video_id = ? AND timestamp BETWEEN ? AND ? ORDER BY timestamp`
    err := r.db.SelectContext(ctx, &stats, query, videoID, from, to)
    return stats, err
}

func (r *sqliteRepository) GetLatestStats(ctx context.Context, videoID string) (*VideoStats, error) {
    var stats VideoStats
    query := `SELECT * FROM video_stats WHERE video_id = ? ORDER BY timestamp DESC LIMIT 1`
    err := r.db.GetContext(ctx, &stats, query, videoID)
    if err == sql.ErrNoRows {
        return nil, nil
    }
    return &stats, err
}

func migrateDatabase(db *sqlx.DB) error {
    // Check if new columns exist, if not add them
    alterQueries := []string{
        `ALTER TABLE video_stats ADD COLUMN comments INTEGER DEFAULT 0`,
        `ALTER TABLE video_stats ADD COLUMN caption TEXT`,
        `ALTER TABLE video_stats ADD COLUMN permalink TEXT`,
        `ALTER TABLE video_stats ADD COLUMN media_type TEXT`,
        `ALTER TABLE video_stats ADD COLUMN media_url TEXT`,
    }

    for _, query := range alterQueries {
        _, err := db.Exec(query)
        // It's okay if the column already exists (will get error)
        if err != nil && !contains(err.Error(), "duplicate column name") {
            fmt.Printf("Migration warning: %v\n", err)
        }
    }

    // Update existing rows to have default values for new columns
    updateQuery := `UPDATE video_stats SET comments = 0 WHERE comments IS NULL`
    _, err := db.Exec(updateQuery)
    if err != nil {
        fmt.Printf("Migration update warning: %v\n", err)
    }
    return nil
}

func contains(s, substr string) bool {
    return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && (s[0:len(substr)] == substr || contains(s[1:], substr)))
}