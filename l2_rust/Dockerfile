FROM rust:1.77-slim

WORKDIR /app

# Install dependencies untuk build
RUN apt-get update && apt-get install -y \
    pkg-config \
    libssl-dev \
    && rm -rf /var/lib/apt/lists/*

# Copy source code
COPY . .

# Build project
RUN cargo build --release

EXPOSE 8081

CMD ["/app/target/release/nusa-vm"]