FROM almalinux:8 as builder
RUN dnf -y install dnf-plugins-core
RUN dnf config-manager --set-enabled powertools
RUN dnf -y install gcc libpcap-devel
RUN curl --proto '=https' --tlsv1.2 -sSf https://sh.rustup.rs | sh -s -- -y
ENV PATH=/root/.cargo/bin:$PATH
WORKDIR /src
COPY . .
RUN cargo build --release

FROM almalinux/8-minimal
RUN microdnf -y install libpcap
COPY --from=builder /src/target/release/dumpy /bin/dumpy
WORKDIR /config
VOLUME /config
ENTRYPOINT ["/bin/dumpy"]
