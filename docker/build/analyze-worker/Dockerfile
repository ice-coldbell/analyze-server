ARG GOLANG_VERSION=1.20

FROM golang:${GOLANG_VERSION}-alpine AS builder

RUN apk update
RUN apk add --no-cache alpine-sdk git

WORKDIR /app

COPY go.mod .
COPY go.sum .
RUN go mod download && go mod verify

COPY application application
COPY pkg pkg
COPY core core

RUN go build -o ./analyze-worker ./application/analyze-worker

################################################################################

FROM alpine
WORKDIR /app

RUN apk update
RUN apk add --no-cache tzdata

RUN cp /usr/share/zoneinfo/Asia/Seoul /etc/localtime
RUN echo "Asia/Seoul" > /etc/timezone

COPY --from=builder /app/analyze-worker analyze-worker

ENTRYPOINT ["/app/analyze-worker"]
