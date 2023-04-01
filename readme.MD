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
