FROM rust:latest

COPY . .

RUN cargo build --release

ENV RUST_LOG=debug

RUN touch .env

CMD ["target/release/api"]