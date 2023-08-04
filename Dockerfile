FROM alpine

RUN touch /var/log/messages

RUN mkdir -p /www/map
RUN chmod a+x /www/map

COPY patches /patches

RUN touch /var/log/messages

RUN apk add --no-cache \
    bash \
    curl \
    zlib \
    lzo \
    openssl \
    iproute2 \
    rsyslog \
    dnsmasq \
    jq \
    gettext \
    wireguard-tools \
    nginx \
    nodejs \
    npm \
    git \
    s6

# Install API dependencies
COPY api /api
RUN cd /api \
    && npm ci

# Install MeshMap dependencies
RUN git clone https://github.com/USA-RedDragon/MeshMap.git /meshmap \
    && cd /meshmap \
    && npm ci

# Workaround for Node 16
ENV NODE_OPTIONS=--openssl-legacy-provider

RUN sed -i 's/module(load="imklog")//g' /etc/rsyslog.conf

# Build and install olsrd
RUN apk add --virtual .olsrd-build-deps \
      git \
      build-base \
      linux-headers \
      bison \
      flex \
    && git clone https://github.com/OLSR/olsrd.git \
    && cd olsrd \
    && git checkout v0.9.8 \
    && for patch in /patches/olsrd/*.patch; do echo "Applying patch: $patch" ; patch -p1 < $patch; done \
    && make prefix=/usr \
    && make prefix=/usr install arprefresh_install txtinfo_install jsoninfo_install dot_draw_install watchdog_install nameservice_install \
    && cd .. \
    && rm -rf olsrd \
    && apk del .olsrd-build-deps \
    && rm -rf /tmp/* /var/cache/apk/*

# Build and install vtun
RUN apk add --virtual .vtun-build-deps \
      build-base \
      linux-headers \
      bison \
      flex \
      zlib-dev \
      lzo-dev \
      binutils \
      openssl-dev \
    && curl -fSsL https://downloads.sourceforge.net/project/vtun/vtun/3.0.3/vtun-3.0.3.tar.gz -o vtun-3.0.3.tar.gz \
    && tar -xzf vtun-3.0.3.tar.gz \
    && rm vtun-3.0.3.tar.gz \
    && cd vtun-3.0.3 \
    # --build=unknown-unknown-linux is magic for cross-compiling
    && ./configure --prefix=/usr --build=unknown-unknown-linux \
    && for patch in /patches/vtun/*.patch; do patch -p1 < $patch; done \
    && make \
    && make install \
    && cd .. \
    && rm -rf vtun-3.0.3 \
    && apk del .vtun-build-deps \
    && rm -rf /tmp/* /var/cache/apk/*

RUN rm -rf /patches

RUN curl -fSsL https://raw.githubusercontent.com/aredn/aredn_packages/3.22.12.0/blockknownencryption/files/20-blockknownencryption -o /usr/bin/blockknownencryption \
    && chmod +x /usr/bin/blockknownencryption

COPY --chown=root:root rootfs /

# Expose ports.
EXPOSE 5525

# Define default command.
CMD ["bash", "/usr/bin/start.sh"]
