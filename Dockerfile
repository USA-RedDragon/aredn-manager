FROM node:22.16.0-alpine AS frontend-build

WORKDIR /app

COPY frontend/package.json frontend/package-lock.json ./
RUN npm ci --ignore-scripts

COPY frontend/. .

ENV NODE_ENV=production

RUN npm run build -- --base=/a/

FROM node:22.16.0-alpine AS new-frontend-build

WORKDIR /app

COPY new-frontend/package.json new-frontend/package-lock.json ./
RUN npm ci --ignore-scripts

COPY new-frontend/. .

ENV NODE_ENV=production

RUN npm run build -- --base=/b/

FROM ghcr.io/usa-reddragon/aredn-base:main@sha256:2064c894a446254f5c084025e8705b15d57c23e9cec4d9e3b086309653d721c9

COPY --from=frontend-build /app/dist /www
COPY --from=new-frontend-build /app/dist /new-www

RUN apk add --no-cache \
    nginx \
    socat

COPY --chown=root:root docker/rootfs/. /

# AREDN Manager runs OLSRD on its own
RUN rm -rf /etc/s6/olsrd

COPY aredn-manager /usr/bin/aredn-manager
CMD ["bash", "/usr/bin/start.sh"]
