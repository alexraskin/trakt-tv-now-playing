# Trakt.tv Now Playing API

Simple API to check what a Trakt.tv user is currently watching.

[![trakt.tv Now Playing](https://img.shields.io/endpoint?color=blueviolet&url=https://trakt.alexraskin.com/alexraskin?format=shields.io)](https://trakt.alexraskin.com/alexraskin)

## Setup

1. Set environment variables:
```bash
TRAKT_CLIENT_ID=your_client_id
TRAKT_CLIENT_SECRET=your_client_secret
ADMIN_KEY=your_admin_key
```

2. Deploy:

### Railway
1. Fork this repository
2. Create new Railway project
3. Add environment variables in Railway dashboard
4. Create a volume in Railway:
   ```bash
   railway volume create data
   ```
5. Deploy! The token file will persist in the `/data` volume

### Docker
```bash
# Build the image
docker build -t trakt-now-playing .

# Run locally with persistent storage
docker run -d \
  -p 8080:8080 \
  -e TRAKT_CLIENT_ID=your_client_id \
  -e TRAKT_CLIENT_SECRET=your_client_secret \
  -e ADMIN_KEY=your_admin_key \
  -v trakt-data:/data \
  --name trakt-now-playing \
  trakt-now-playing
```

### Local Development
```bash
go run main.go
```

## Endpoints

- `GET /:username` - Check what user is watching
- `GET /admin/auth` - Authorize the application
- `GET /admin/status` - Check auth status
- `GET /admin/refresh` - Refresh the token 