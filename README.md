# YouTube & Instagram Stats Tracker

A Go-based service that tracks video & posta statistics (likes & comments) from YouTube and Instagram platforms respectively. This application periodically polls metrics and stores historical data for analysis.

## Overview

The **YouTube-IG-stats-tracker** is a microservice designed to:

- Register videos from YouTube and Instagram for tracking
- Automatically poll and collect video statistics at regular intervals
- Store historical statistics in a database
- Provide REST APIs to query video performance metrics

## Tech Stack

- **Language**: Go 1.24+
- **Web Framework**: Gorilla Mux
- **HTTP Transport**: go-kit
- **Database**: SQLite (with PostgreSQL support)
- **Container**: Docker & Docker Compose
- **Task Scheduler**: Robfig Cron

## Features

### Core Functionality

- **Video Registration**: Register YouTube or Instagram videos for tracking
- **Automatic Polling**: Periodic background workers collect video stats
- **Statistics Storage**: Historical data stored for trend analysis
- **REST API**: Query video statistics within date ranges

### Supported Platforms

- **YouTube**: Full integration with YouTube API v3
- **Instagram**: Support for Instagram Graph API

## API Endpoints

### Track Video

```
POST /track-video
Content-Type: application/json

{
  "platform": "youtube|instagram",
  "video_id": "string",
  "username": "string (optional)",
  "tag": "string (optional)"
}
```

### Get Statistics

```
GET /stats?video_id=<id>&from=<timestamp>&to=<timestamp>
```

Returns video statistics for the specified date range.

## Prerequisites

- Docker & Docker Compose (recommended)
- Go 1.24+ (for local development)
- YouTube API Key
- Instagram API Token

## Environment Variables

Configure the following environment variables:

| Variable          | Description                   | Default          |
| ----------------- | ----------------------------- | ---------------- |
| `YOUTUBE_API_KEY` | YouTube API v3 key            | (required)       |
| `INSTAGRAM_TOKEN` | Instagram Graph API token     | (required)       |
| `INSTAGRAM_ID`    | Instagram business account ID | (required)       |
| `HTTP_PORT`       | HTTP server port              | `8080`           |
| `DATABASE_URL`    | Database connection string    | SQLite in-memory |

## Getting Started

### Using Docker Compose (Recommended)

1. **Clone the repository**

   ```bash
   git clone https://github.com/divakarrajput1/YouTube-IG-stats-tracker.git
   cd video-stats-tracker
   ```

2. **Set up environment variables**
   Create a `.env` file or update `docker-compose.yml`:

   ```bash
   YOUTUBE_API_KEY=your_youtube_api_key
   INSTAGRAM_TOKEN=your_instagram_token
   INSTAGRAM_ID=your_instagram_business_id
   ```

3. **Start the services**

   ```bash
   docker-compose up -d
   ```

   The application will be available at `http://localhost:8080`

### Local Development

1. **Run the application**

   ```bash
   export YOUTUBE_API_KEY=your_key
   export INSTAGRAM_TOKEN=your_token
   export INSTAGRAM_ID=your_id

   go run ./cmd/server/main.go
   ```
## Configuration
```

### API Keys

1. **YouTube API**:

   - Visit [Google Cloud Console](https://console.cloud.google.com/)
   - Create a new project
   - Enable YouTube Data API v3
   - Generate an API key

2. **Instagram**:
   - Create a Meta App at [Meta Developers](https://developers.facebook.com/)
   - Set up Instagram Graph API
   - Generate access token

## Usage Examples

### Track a YouTube Video

```bash
curl -X POST http://localhost:8080/track-video \
  -H "Content-Type: application/json" \
  -d '{
    "platform": "youtube",
    "video_id": "dQw4w9WgXcQ",
    "username": "example_user",
    "tag": "tutorial"
  }'
```

## Architecture

1. **Transport Layer** (`internal/transport/http/`)

   - Handles HTTP request/response serialization
   - Routes incoming requests

2. **Endpoint Layer** (`internal/endpoint/`)

   - go-kit endpoint definitions
   - Request validation and response formatting

3. **Service Layer** (`internal/service/`)

   - Core business logic
   - Coordinates between repository and platform clients

4. **Platform Layer** (`internal/platform/`)

   - YouTube and Instagram API integrations
   - Handles API-specific operations

5. **Repository Layer** (`internal/repository/`)

   - Database abstraction
   - CRUD operations for videos and statistics

6. **Worker Layer** (`internal/worker/`)
   - Background polling tasks
   - Scheduled statistics collection

## Polling Strategy

The application uses Robfig Cron to schedule periodic polling:

- Polls are executed at configurable intervals
- Each poll fetches current statistics from the respective platform
- Statistics are stored with timestamps for historical tracking
