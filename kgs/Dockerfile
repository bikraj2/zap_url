FROM golang:1.23.1-alpine AS builder

WORKDIR /app 


RUN adduser -D -u  1001 nonroot

RUN apk add --no-cache gcc musl-dev

# Enable CGO for external linking
ENV CGO_ENABLED=1

COPY ../go.mod ../go.sum ./

RUN go mod download

copy . .


RUN go build -o ./main \
    -ldflags="-linkmode external -extldflags -static" \
    ./kgs/cmd/


FROM alpine:latest

COPY --from=builder /app/main kgs 
RUN apk add --no-cache libc6-compat


CMD ["/kgs"]
