FROM golang:1.23.3-alpine as build

WORKDIR /build
LABEL org.opencontainers.image.source="https://github.com/simbafs/coscup-attendance"
LABEL org.opencontainers.image.authors="Simba Fs <me@simbafs.cc>"

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -ldflags "-s -w" -o bot

FROM scratch 

WORKDIR /app

COPY --from=build /build/bot /app/bot
COPY --from=build /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

CMD [ "./bot" ]
