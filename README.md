# Trakt.tv Now Playing API

Simple API to check what a Trakt.tv user is currently watching.

[![Image from Gyazo](https://i.gyazo.com/034ad27b9ebe8665cc5bce96b264ce65.png)](https://gyazo.com/034ad27b9ebe8665cc5bce96b264ce65)

[![trakt.tv Now Playing](https://img.shields.io/endpoint?url=https://trakt.alexraskin.com/twizycat?format=shields.io)](https://github.com/alexraskin/trakt-tv-now-playing)


## Running the API

Get your Trakt.tv API key from [Trakt.tv](https://trakt.tv/oauth/applications)

### Docker
```bash
# Build the image
docker build -t trakt-now-playing .

# Run the container
docker run -d \
  -p 4000:4000 \
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
