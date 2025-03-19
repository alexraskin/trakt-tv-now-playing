# Trakt.tv Now Playing API

Simple API to check what a Trakt.tv user is currently watching.

```bash
curl -s https://trakt.alexraskin.com/{your trakt username} | jq .
```

[![trakt.tv Now Playing](https://img.shields.io/endpoint?color=blueviolet&url=https://trakt.alexraskin.com/alexraskin?format=shields.io)](https://trakt.alexraskin.com/alexraskin)

## Setup

1. Set environment variables:
```bash
TRAKT_CLIENT_ID=your_client_id
TRAKT_CLIENT_SECRET=your_client_secret
ADMIN_KEY=your_admin_key
```

2. Run with Docker:
```bash
docker build -t trakt-now-playing .
docker run -p 8080:8080 --env-file .env trakt-now-playing
```

Or run locally:
```bash
go run main.go
```

## Endpoints

- `GET /:username` - Check what user is watching
- `GET /admin/auth` - Authorize the application
- `GET /admin/status` - Check auth status
- `GET /admin/refresh` - Refresh the token 