FROM golang:1.21.4

# Update the package repository and install ffmpeg
RUN apt-get update && \
    apt-get install -y ffmpeg

# Set the Current Working Directory inside the container
WORKDIR /app


COPY ./challenge .

RUN go get

# Copy the source from the current directory to the Working Directory inside the container
COPY ./challenge .
COPY ./flag.txt ./flag.txt

# Build the Go app
RUN go build -o main .

# Expose port 8080 to the outside world
EXPOSE 1337

# Rename ./flag.txt to ./flag_<32 random characters>.txt
RUN mv ./flag.txt ./flag_$(head /dev/urandom | tr -dc A-Za-z0-9 | head -c 32).txt

# Command to run the executable
CMD ["./main"]