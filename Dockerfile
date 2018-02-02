# use builder image to compile ledisdb (without GCO)
FROM golang:1.9-stretch as builder

ENV LEDISDB_VERSION 0.6

ENV LEVELDB_VERSION 47cb9e2a211e1d7157078ba7bab536beb29e56dc
ENV ROCKSDB_VERSION 5.8.6
ENV GOSU_VERSION 1.10



WORKDIR /build

RUN apt-get update && \
    apt-get install -y \
    ca-certificates \
    wget \
    gcc-6 \
    g++-6 \
    build-essential \
    libsnappy1v5 \
    libsnappy-dev \
    libgflags-dev

# get LedisDB
RUN wget -O ledisdb-src.tar.gz "https://github.com/siddontang/ledisdb/archive/v$LEDISDB_VERSION.tar.gz" && \
    tar -zxf ledisdb-src.tar.gz && \
    mkdir -p /go/src/github.com/siddontang/ && \
    mv ledisdb-$LEDISDB_VERSION /go/src/github.com/siddontang/ledisdb

# build LevelDB
RUN wget -O leveldb-src.tar.gz "https://github.com/google/leveldb/archive/$LEVELDB_VERSION.tar.gz" && \
    tar -zxf leveldb-src.tar.gz && \
    cd leveldb-$LEVELDB_VERSION && \
    patch -p0 < /go/src/github.com/siddontang/ledisdb/tools/leveldb.patch && \
    mkdir -p out-shared/db && \
    make -j "$(nproc)" && \
    mkdir /build/lib && \
    mkdir -p /build/include/leveldb && \
    cp out-static/lib* /build/lib/ && \
    install include/leveldb/*.h /build/include/leveldb


# build RocksDB
RUN wget -O rocksdb-src.tar.gz "https://github.com/facebook/rocksdb/archive/v$ROCKSDB_VERSION.tar.gz" && \
    tar -zxf rocksdb-src.tar.gz && \
    cd rocksdb-$ROCKSDB_VERSION && \
    make static_lib -j "$(nproc)" && \
    mkdir -p /build/include/rocksdb && \
    cp librocksdb.a /build/lib/ && \
    install include/rocksdb/*.h /build/include/rocksdb

ENV CGO_CFLAGS "-I/build/include"
ENV CGO_CXXFLAGS "-I/build/include"
ENV CGO_LDFLAGS "-L/build/lib -lsnappy"

#build LedisDB
RUN export LD_LIBRARY_PATH=$LD_LIBRARY_PATH:/build/lib:/usr/lib && \
    export DYLD_LIBRARY_PATH=$DYLD_LIBRARY_PATH:/build/lib:/usr/lib && \
    mkdir -p /build/bin && \
    rm -rf /build/bin/* && \
    cd /go/src/github.com/siddontang/ledisdb && \
    GOGC=off go build -i -o /build/bin/ledis-server -tags "snappy leveldb rocksdb" cmd/ledis-server/* && \
    GOGC=off go build -i -o /build/bin/ledis-cli -tags "snappy leveldb rocksdb" cmd/ledis-cli/* && \
    GOGC=off go build -i -o /build/bin/ledis-benchmark -tags "snappy leveldb rocksdb" cmd/ledis-benchmark/* && \
    GOGC=off go build -i -o /build/bin/ledis-dump -tags "snappy leveldb rocksdb" cmd/ledis-dump/* && \
    GOGC=off go build -i -o /build/bin/ledis-load -tags "snappy leveldb rocksdb" cmd/ledis-load/* && \
    GOGC=off go build -i -o /build/bin/ledis-repair -tags "snappy leveldb rocksdb" cmd/ledis-repair/*

# grab gosu for easy step-down from root
# https://github.com/tianon/gosu/releases
RUN set -ex; \
    dpkgArch="$(dpkg --print-architecture | awk -F- '{ print $NF }')"; \
    wget -O /usr/local/bin/gosu "https://github.com/tianon/gosu/releases/download/$GOSU_VERSION/gosu-$dpkgArch"; \
    wget -O /usr/local/bin/gosu.asc "https://github.com/tianon/gosu/releases/download/$GOSU_VERSION/gosu-$dpkgArch.asc"; \
    export GNUPGHOME="$(mktemp -d)"; \
    gpg --keyserver ha.pool.sks-keyservers.net --recv-keys B42F6819007F00F88E364FD4036A9C25BF357DD4; \
    gpg --batch --verify /usr/local/bin/gosu.asc /usr/local/bin/gosu; \
    chmod +x /usr/local/bin/gosu


# done building - now create a tiny image with a statically linked Ledis
FROM debian:stretch-slim

COPY --from=builder /build/lib/* /usr/lib/
COPY --from=builder /build/bin/ledis-* /bin/
COPY --from=builder /go/src/github.com/siddontang/ledisdb/config/config-docker.toml /config.toml
COPY --from=builder /usr/local/bin/gosu /bin/

RUN groupadd -r ledis && \
    useradd -r -g ledis ledis && \
    mkdir /datastore && \
    chown ledis:ledis /datastore && \
    chmod 444 /config.toml && \
    gosu nobody true

RUN apt-get update && \
    apt-get install -y --no-install-recommends libsnappy1v5 && \
    rm -rf /var/lib/apt/lists/*

VOLUME /datastore

ADD entrypoint.sh /bin/entrypoint.sh

ENTRYPOINT ["entrypoint.sh"]

EXPOSE 6380 11181

CMD ["ledis-server", "--config=/config.toml"]
