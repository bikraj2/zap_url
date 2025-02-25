FROM golang:1.23.1-alpine AS builder

WORKDIR /app 


RUN adduser -D -u  1001 nonroot

RUN apk add --no-cache gcc musl-dev

# Enable CGO for external linking
ENV CGO_ENABLED=1

COPY ../go.mod ../go.sum ./

COPY .envrc /envrc
RUN go mod download
copy . .


RUN go build -o ./main \
    ./gateway/cmd/


FROM alpine:latest

COPY --from=builder /app/main /shorten

COPY .envrc /envrc

# Ensure the env file is sourced before running
RUN echo 'source /envrc' >> /etc/profile

EXPOSE 8084

CMD ["sh", "-c", "source /envrc && /shorten -cors-trusted-origin \"$CORS_ORIGIN\""]

