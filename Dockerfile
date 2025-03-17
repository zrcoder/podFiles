FROM golang:1.24-alpine AS builder

WORKDIR /app

COPY . .

ENV GOPROXY=https://goproxy.cn,direct

ARG USE_LOCAL_SDK=false

RUN if [ "$USE_LOCAL_SDK" = "true" ]; then \
        go build -tags local_sdk -ldflags "-s -w" -o podfiles ./cmd/podFiles; \
    else \
        go build -ldflags "-s -w" -o podfiles ./cmd/podFiles; \
    fi

FROM alpine:3.21.3

RUN adduser -D -g '' appuser

WORKDIR /app

COPY --from=builder /app/podfiles .

USER appuser

EXPOSE 8080

ENTRYPOINT ["./podfiles"]