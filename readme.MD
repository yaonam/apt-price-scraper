# Apartment Price Scraper

## Overview

This package makes it easier to check the latest apartment pricing from Proto Kendall Square. Sends messages to Discord channel whenever prices are updated and prints to log.

## Installation

Git clone: https://github.com/yaonam/apt-price-scraper.git
(Use vim to create env var)

vim: https://www.cyberciti.biz/faq/vim-new-file-creation-command-on-linux-unix/

Docker: https://docs.docker.com/engine/install/ubuntu/#install-using-the-repository

Go: https://go.dev/doc/install

## Run

### Locally

```
cp .env.example .env // Make sure to fill out env file
go run .
```

### Server - Docker Compose
```
// Start
docker-compose up --build -d

// Shut down
docker-compose stop
docker-compose down
```

### Server - Docker Build
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
docker rm <container id>

// Check logs
docker logs <container id>
```

## Develop

### Docker

https://docs.docker.com/language/golang/build-images/
