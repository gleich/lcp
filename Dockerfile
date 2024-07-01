FROM rust:1.79.0

COPY . .

RUN cargo build --release

ENV RUST_LOG=debug

RUN touch .env

CMD ["target/release/lcp"]