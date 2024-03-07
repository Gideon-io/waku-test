FROM ubuntu:latest

# Install required dependencies
RUN apt-get update && apt-get install -y \
    wget \
    tar \
    gzip \
    ca-certificates

# Download and install Go 1.2
RUN wget https://go.dev/dl/go1.20.14.linux-amd64.tar.gz && \
    tar -C /usr/local -xzf go1.20.14.linux-amd64.tar.gz
    

# Set Go environment variables
ENV PATH="/usr/local/go/bin:${PATH}"
ENV GOPATH="/go"

# Create a directory for your Go code
RUN mkdir /go

# Set the working directory
WORKDIR /go

# Copy your Go code into the container
COPY . .

# Build and run your Go application
RUN go build -o app .