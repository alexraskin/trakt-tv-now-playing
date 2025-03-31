FROM --platform=$BUILDPLATFORM golang:1.24-alpine AS build

WORKDIR /build

COPY go.mod go.sum ./
RUN go mod download

COPY . .

ARG TARGETOS
ARG TARGETARCH
ARG VERSION
ARG COMMIT
ARG BUILD_TIME

RUN --mount=type=cache,target=/root/.cache/go-build \
    --mount=type=cache,target=/go/pkg \
    CGO_ENABLED=0 \
    GOOS=$TARGETOS \
    GOARCH=$TARGETARCH \
    go build -ldflags="-X 'main.version=$VERSION' -X 'main.commit=$COMMIT' -X 'main.buildTime=$BUILD_TIME'" -o trakt-tv-now-playing github.com/alexraskin/trakt-tv-now-playing

FROM alpine

RUN apk --no-cache add ca-certificates

COPY --from=build /build/trakt-tv-now-playing /bin/trakt-tv-now-playing

EXPOSE 4000

ENTRYPOINT ["/bin/trakt-tv-now-playing"]

CMD ["-port", "4000"]