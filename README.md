# KonaIndex

A high-performance image search thingy for [Konachan.net](https://konachan.net). Advanced color-based and tag-based image search using vector similarity

## Overview

KonaIndex indexes artwork from Konachan.net and provides a powerful search API that allows you to find images by:
- **Color palette** - Search by up to 3 RGB colors (converted to LAB later)
- **Tags** - Include or exclude specific tags
- **Ratings** - Filter by safe (s), explicit (e), or questionable (q)

## Prerequisites

- **Go** 1.25.7 or higher
- **PostgreSQL** with pgvector extension


### Environment Configuration

Create a `.env` file in the project root:

```env
DATABASE_URL=host=localhost user=postgres password=postgres dbname=konaindex port=5432 sslmode=disable TimeZone=UTC
SERVER_PORT=8000
WORKER_COUNT=5
HISTORICAL_HARD_CAP_ID=140000
```


## API Usage

### Search Endpoint

**POST** `/api/search`

```json
{
  "include_tags": ["landscape", "sunset"],
  "exclude_tags": ["character"],
  "limit": 20,
  "target_colors": [
    [255, 100, 50],
    [30, 144, 255]
  ],
  "ratings": ["s"]
}
```

**Request Parameters:**
- `include_tags` - Array of tags that must be present (uses PostgreSQL array contains `@>`)
- `exclude_tags` - Array of tags to exclude (uses PostgreSQL array overlap `&&`)
- `limit` - Maximum number of results (default: 20)
- `target_colors` - Up to 3 RGB colors `[R, G, B]` (0-255)
- `ratings` - Array of rating filters: `"s"` (safe), `"e"` (explicit), `"q"` (questionable)

**Response:**

```json
[
  {
    "ID": 123,
    "CreatedAt": "2026-03-07T16:06:38Z",
    "UpdatedAt": "2026-03-07T16:21:34Z",
    "DeletedAt": null,
    "KonachanID": 118191,
    "ImageURL": "https://konachan.net/image/...",
    "PreviewURL": "https://konachan.net/data/preview/...",
    "Tags": ["landscape", "sunset", "ocean"],
    "Rating": "s",
    "Temperature": 3052.58,
    "Palette": [
      {
        "ID": 2001,
        "PostID": 123,
        "Color": [0.558, 0.371, 0.440],
        "Weight": 0.233
      }
    ]
  }
]
```


### Sync Worker
- Runs every hour
- Fetches the latest 100 posts from Konachan
- Inserts new posts and queues them for processing

### Catchup Worker
- Runs continuously
- Backfills historical posts in descending order
- Stops when reaching `HISTORICAL_HARD_CAP_ID`
- Fetches max 100 posts per batch

### Palette Workers
- Configurable worker pool (default: 5 workers)
- Downloads preview images
- Extracts 5 dominant colors using K-means clustering
- Calculates color temperature (CCT)
- Stores colors as LAB vectors in PostgreSQL
- Boosts weight for highly saturated colors (chroma > 40)

## Acknowledgments

- [Konachan.net](https://konachan.net) for providing the image API
- [pgvector](https://github.com/pgvector/pgvector) for vector similarity search
- [go-colorful](https://github.com/lucasb-eyer/go-colorful) for color space conversions


hehe