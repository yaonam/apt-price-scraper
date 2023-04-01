# Apartment Price Scraper

## Overview

This package makes it easier to check the latest apartment pricing from Proto Kendall Square.

## Run

```
cp .env.example .env // Make sure to fill out env file
go run .
```

## Server
Move `server.service` file to /etc/systemd/system/
```
sudo systemctl start server
```

## Docker

https://docs.docker.com/language/golang/build-images/

### Docker Compose
```
// Start
docker-compose up --build -d

// Shut down
docker-compose stop
docker-compose down
```

### Docker Build
```
// Build the image
docker build --tag scraper

// Run the image in detached mode
docker run -d --name apt-scraper scraper
```

### Misc
```
// List images
docker images

// List containers
docker ps

// Remove image
docker rmi <image name>

// Remove containers
docker rm <container name>
```
