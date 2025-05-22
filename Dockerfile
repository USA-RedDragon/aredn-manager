FROM node:22.16.0-alpine AS frontend-build

WORKDIR /app

COPY frontend/package.json frontend/package-lock.json ./
RUN npm ci --ignore-scripts

COPY frontend/. .

ENV NODE_ENV=production

RUN npm run build

FROM ghcr.io/usa-reddragon/aredn-base:main@sha256:00d5f050e48f5a265522dc7d47db3eedd1eefce8680f726518cafa4dbf2f233d

COPY --from=frontend-build /app/dist /www

RUN apk add --no-cache \
    nginx \
    socat

COPY --chown=root:root docker/rootfs/. /

# AREDN Manager runs OLSRD on its own
RUN rm -rf /etc/s6/olsrd

COPY aredn-manager /usr/bin/aredn-manager
CMD ["bash", "/usr/bin/start.sh"]
