FROM ubuntu as vtun-builder

RUN find /etc/apt/sources.list* -type f -exec sed -i 'p; s/^deb /deb-src /' '{}' +

# Compile vtun because Ubuntu's doesn't compile with HAVE_WORKING_FORK, causing breakages
RUN DEBIAN_FRONTEND=noninteractive apt-get update \
    && DEBIAN_FRONTEND=noninteractive apt-get install -y --no-install-recommends \
        build-essential \
        fakeroot \
        dpkg-dev \
        devscripts \
    && DEBIAN_FRONTEND=noninteractive apt-get build-dep -y vtun \
    && DEBIAN_FRONTEND=noninteractive apt-get update \
    && DEBIAN_FRONTEND=noninteractive apt-get source --tar-only --diff-only vtun \
    && tar xf vtun_*.orig.tar.* \
    && cd vtun-* \
    && tar xf ../vtun_*.debian.tar.* \
    && rm debian/patches/00-sslauth.patch \
    && sed -i -e '1d' debian/patches/series \
    && sed -i -e '10,46d' debian/patches/03-signedness-warnings.patch \
    && sed -i -e 's/ K_SSLAUTH//' debian/patches/07-dual-family-transport.patch \
    && sed -i -e '39d' debian/patches/07-dual-family-transport.patch \
    && sed -i -e '36 a\\      vtun.cfg_file = VTUN_CONFIG_FILE;' debian/patches/07-dual-family-transport.patch \
    && debuild -us -uc -i -I

FROM ubuntu

COPY --from=vtun-builder /vtun_3.0.4-2build1_amd64.deb /vtun_3.0.4-2build1_amd64.deb

RUN DEBIAN_FRONTEND=noninteractive apt-get update \
    && DEBIAN_FRONTEND=noninteractive apt-get install -y \
        iproute2 \
        /vtun_3.0.4-2build1_amd64.deb \
    && rm -rf /var/lib/apt/lists/* /vtun_3.0.4-2build1_amd64.deb

COPY --chown=root:root rootfs /

# Expose ports.
EXPOSE 5525

# Define default command.
CMD ["bash", "/usr/bin/start.sh"]
