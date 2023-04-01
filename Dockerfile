FROM golang:1.20-alpine

WORKDIR /app

# Copy the module files into working dir
COPY go.mod ./
COPY go.sum ./

RUN go mod download

# Copy all go files into dir
COPY *.go ./

# Build into the image root
RUN go build -o /apt-price-scraper

CMD [ "/apt-price-scraper" ]