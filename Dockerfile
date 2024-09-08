# Start with Ubuntu base image
FROM ubuntu:22.04

# Set environment variables
ENV DEBIAN_FRONTEND=noninteractive

# Update and install necessary packages
RUN apt-get update && apt-get install -y \
    curl \
    gcc \
    make \
    llvm \
    clang \
    libbpf-dev \
    linux-headers-generic \
    && rm -rf /var/lib/apt/lists/*

RUN arch=$(uname -m) && \
    case $arch in \
        x86_64|x86) ln -s /usr/include/x86_64-linux-gnu/asm /usr/include/asm ;; \
        aarch64|arm64) ln -s /usr/include/aarch64-linux-gnu/asm /usr/include/asm ;; \
        i386|i686) ln -s /usr/include/i386-linux-gnu/asm /usr/include/asm ;; \
        armv7l|armv6l) ln -s /usr/include/arm-linux-gnueabihf/asm /usr/include/asm ;; \
        *) echo "Unsupported architecture: $arch" && exit 1 /usr/include/asm ;; \
    esac

# Download and install Go
RUN curl -fsSL https://golang.org/dl/go1.23.1.linux-amd64.tar.gz -o go.tar.gz \
    && tar -C /usr/local -xzf go.tar.gz \
    && rm go.tar.gz

# Set up Go environment variables
ENV PATH=$PATH:/usr/local/go/bin

# Create and set the working directory
WORKDIR /app
COPY . .

# Run make command to generate binaries
RUN make nw-rx-tracker
