FROM golang:1.21-alpine AS aredn-manager

WORKDIR /app

RUN apk add --no-cache git bash

COPY manager/go.mod manager/go.sum ./
RUN go mod download

COPY manager/. .
COPY .git .git

RUN [ -f ./internal/sdk/commit.txt ] || go generate ./...

RUN go build -o aredn-manager ./main.go

FROM node:18-alpine AS aredn-manager-frontend

WORKDIR /app

COPY manager/frontend/package.json manager/frontend/package-lock.json ./
RUN npm ci

COPY manager/frontend/. .

ENV NODE_ENV=production

RUN npm run build

FROM ghcr.io/usa-reddragon/aredn-base:next

COPY --from=aredn-manager /app/aredn-manager /usr/bin/aredn-manager
RUN chmod a+x /usr/bin/aredn-manager

COPY --from=aredn-manager-frontend /app/dist /www/aredn-manager

RUN apk add --no-cache \
    nginx

# Install API dependencies
COPY api /api
RUN cd /api \
    && npm ci

COPY --chown=root:root rootfs /

# Expose ports.
EXPOSE 5525

# Define default command.
CMD ["bash", "/usr/bin/start.sh"]
