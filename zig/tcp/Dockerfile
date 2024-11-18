# Use a base image that suits your binary's dependencies
FROM alpine:latest

# Set up working directory
WORKDIR /app

# Copy your Linux binary into the container
COPY main /app/

# Make the binary executable
RUN chmod +x /app/main
EXPOSE 5882 

# Specify the command to run your binary
CMD ["./main"]
