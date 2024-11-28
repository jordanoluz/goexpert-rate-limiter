FROM golang:1.23.2 AS build

WORKDIR /app

COPY . .

RUN go mod tidy && go test ./... -v
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o rate-limiter ./cmd/api

FROM scratch

WORKDIR /app

COPY --from=build /app/rate-limiter .
COPY --from=build /app/.env .

EXPOSE 8080

ENTRYPOINT [ "./rate-limiter" ]