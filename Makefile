# Variables
BINARY_NAME=thdctl
DOCKER_IMAGE=thdctl:latest

# Build the Go executable
build:
	go build -o $(BINARY_NAME) .

# Build the Docker image
docker-build:
	docker build -t $(DOCKER_IMAGE) .

# Build both the executable and the Docker image
all: build docker-build

# Clean up the built executable
clean:
	rm -f $(BINARY_NAME)

# Phony targets
.PHONY: build docker-build all clean
