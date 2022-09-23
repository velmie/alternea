FROM golang:1.19-alpine AS builder

RUN apk update && apk add --no-cache git mercurial

WORKDIR $GOPATH/src/velmie/alternea

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -gcflags "all=-N -l" -o ./build/alternea ./cmd/alternea/main.go


FROM surnet/alpine-wkhtmltopdf:3.16.0-0.12.6-full as wkhtmltopdf


FROM alpine:3.16

# Install needed packages
RUN  apk add --no-cache \
      libstdc++ \
      libx11 \
      libxrender \
      libxext \
      libssl1.1 \
      ca-certificates \
      fontconfig \
      freetype \
      ttf-dejavu \
      ttf-droid \
      ttf-freefont \
      ttf-liberation \
      tzdata \
    && apk add --no-cache --virtual .build-deps \
      msttcorefonts-installer \
    \
    # Install microsoft fonts
    && update-ms-fonts \
    && fc-cache -f \
    \
    # Clean up when done
    && rm -rf /tmp/* \
    && apk del .build-deps


RUN addgroup -g 1000 alternea \
    && adduser -D -g "" -h "/app" -G alternea -H -u 1000 alternea

USER alternea

WORKDIR /app

COPY --from=builder /go/src/velmie/alternea/build/alternea /app/alternea
COPY --from=wkhtmltopdf /bin/wkhtmltopdf /usr/bin/wkhtmltopdf
COPY --from=wkhtmltopdf /bin/wkhtmltoimage /usr/bin/wkhtmltoimage

ENTRYPOINT ["/app/alternea"]
