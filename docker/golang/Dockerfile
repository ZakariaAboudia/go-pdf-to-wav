# Use the official Go image as the base image
FROM golang:1.22.6-bullseye

# Set the working directory inside the container
WORKDIR /app

# Copy the Go application source code to the container
COPY . /app

# Install ffmpeg
RUN apt-get update && apt-get install -y ffmpeg

# RUN apt-get -y install poppler-utils wv unrtf tidy

RUN go mod init go-pdf-to-wav

RUN go mod tidy

#RUN go get github.com/JalfResi/justext

# Build the Go application
#RUN cd /app && RUN go build -o main .

# Expose the port the application will run on
EXPOSE 8080

# Command to run the application
# CMD ["go"]

# Keep the container running
CMD tail -f /dev/null

