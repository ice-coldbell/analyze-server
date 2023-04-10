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
COPY internal internal

RUN go build -o ./analyze-server ./application/analyze-server

################################################################################

FROM alpine
WORKDIR /app

RUN apk update
RUN apk add --no-cache tzdata

RUN cp /usr/share/zoneinfo/Asia/Seoul /etc/localtime
RUN echo "Asia/Seoul" > /etc/timezone

COPY --from=builder /app/analyze-server analyze-server

ENTRYPOINT ["/app/analyze-server"]
