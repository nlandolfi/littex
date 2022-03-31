# Use the offical golang image to create a binary.
# This is based on Debian and sets the GOPATH to /go.
# https://hub.docker.com/_/golang
FROM golang:1.17-buster as builder

# Create and change to the app directory.
WORKDIR /app

# Retrieve application dependencies.
# This allows the container build to reuse cached dependencies.
# Expecting to copy go.mod and if present go.sum.
COPY go.* ./
# RUN go mod download # not needed if using prebuilt

# Copy local code to the container image.
COPY . ./

# Build the binary.
COPY ./cmd/server/linuxserver server
# RUN go build -o server -mod readonly -v ./cmd/server/server.go # not needed if using prebuilt


# Use the official Debian slim image for a lean production container.
# https://hub.docker.com/_/debian
# https://docs.docker.com/develop/develop-images/multistage-build/#use-multi-stage-builds
FROM debian:buster-slim
RUN set -x && apt-get update && DEBIAN_FRONTEND=noninteractive apt-get install -y \
    ca-certificates && \
    rm -rf /var/lib/apt/lists/*

# Copy the binary to the production image from the builder stage.
COPY --from=builder /app/server /app/server
# Copy in the static
# below line is for everyone except app
#COPY ./static /static
# below line is for app
COPY ./cmd/server/static /static

# Run the web service on container startup.
CMD ["/app/server"]
