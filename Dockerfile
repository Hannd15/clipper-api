FROM golang:1.22.6 AS build-stage

COPY go.mod go.sum ./
RUN go mod download

COPY *.go ./

RUN CGO_ENABLED=0 GOOS=linux go build -o /ClipperAPI

FROM alpine:3.14 AS final-stage

COPY --from=build-stage /ClipperAPI /ClipperAPI

RUN mkdir -p /videos /clips

RUN apk update && apk upgrade && apk add ffmpeg python3 py-pip
RUN apk add --no-cache gcc musl-dev python3-dev
RUN pip install yt-dlp

ENV PORT="8080"
ENV HOST="localhost"
ENV REMOTE_IP="0.0.0.0"
ENV CORS="*"
EXPOSE $PORT

RUN chmod +x ./ClipperAPI

ENTRYPOINT ["./ClipperAPI"]