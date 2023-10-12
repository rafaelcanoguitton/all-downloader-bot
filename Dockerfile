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
# RUN apt install yt-dlp -y
RUN curl -L https://github.com/yt-dlp/yt-dlp/releases/latest/download/yt-dlp -o /usr/local/bin/yt-dlp
RUN chmod a+rx /usr/local/bin/yt-dlp  # Make executable
CMD ["/all-downloader-bot"]
