FROM golang:latest
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY *.go ./
RUN CGO_ENABLED=0 GOOS=linux go build -o /all-downloader-bot
#Make sure ffmpeg is installed
RUN apt-get update && apt-get install -y ffmpeg
ENV TELEGRAM_TOKEN=""
#make downloads folder
RUN mkdir downloads
CMD ["/all-downloader-bot"]
