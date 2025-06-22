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

FROM ghcr.io/usa-reddragon/aredn-base:main@sha256:78bc39c975673c7c412240ffe596152b68e89354015226ac9ba76f781d40e97a

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
