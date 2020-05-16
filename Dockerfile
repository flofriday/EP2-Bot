# Choose apline because it has the smallest footprint (download size)
FROM golang:1.14.3-alpine

# Install git for go download and tzdata to have to correct time
# Also set the correct time and print it
RUN apk add git tzdata && \
cp /usr/share/zoneinfo/Europe/Berlin /etc/localtime && \
echo "Europe/Berlin" >  /etc/timezone && \
date

# Set the folder of our bot
WORKDIR /app

# Copy the files specifying the dependencies and download them
COPY go.mod go.sum ./
RUN go mod download

# Copy the rest of the files
COPY . .

# Compile the server
RUN GO111MODULE=on CGO_ENABLED=0 go build -ldflags "-s -X 'main.buildDate=$(date)'"

# Run the server
ENTRYPOINT ["./EP2-Bot"]

