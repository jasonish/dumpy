FROM alpine:latest AS builder
RUN apk add rust cargo libpcap-dev
WORKDIR /src
COPY . .
RUN cargo build --release

FROM alpine:latest
RUN apk add libpcap libgcc
COPY --from=builder /src/target/release/dumpy /bin/dumpy
WORKDIR /config
VOLUME /config
ENTRYPOINT ["/bin/dumpy"]
