FROM golang:1.25.5 AS build

WORKDIR /app

# Dependencies
COPY go.mod go.sum ./
RUN go mod download

# Source
COPY . .

# Build
RUN CGO_ENABLED=0 GOOS=linux go build -o /auth

FROM nginx:alpine

RUN apk add --no-cache curl

WORKDIR /app

COPY --from=build /auth /app/auth

EXPOSE 8080

ENTRYPOINT [ "/app/auth" ]
