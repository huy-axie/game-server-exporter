FROM golang:buster AS builder
ENV GO111MODULE=on

WORKDIR /src

COPY . ./

RUN CGO_ENABLED=0 GOOS=linux go build -o svc

FROM debian:buster-slim AS runner

COPY --from=builder /src/svc /svc/
WORKDIR /svc
EXPOSE 9101
ENTRYPOINT ["./svc"]