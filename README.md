# Trakt.tv Now Playing API

Simple API to check what a Trakt.tv user is currently watching.

[![trakt.tv Now Playing](https://img.shields.io/endpoint?url=https://trakt.alexraskin.com/twizycat?format=shields.io)](https://github.com/alexraskin/trakt-tv-now-playing)


## Running the API

Get your Trakt.tv API key from [Trakt.tv](https://trakt.tv/oauth/applications)

### Docker
```bash
# Build the image
docker build -t trakt-now-playing .

# Run the container
docker run -d \
  -p 8080:8080 \
  -e TRAKT_API_KEY=your_client_id \
  --name trakt-now-playing \
  trakt-now-playing
```

### Local Development
```
export TRAKT_API_KEY=your_client_id
go run main.go
```

## Usage

### Check what a user is watching
```
GET /:username
```

### Get shields.io badge
```
GET /:username?format=shields.io
```

Example: `http://localhost:8080/username?format=shields.io` 
