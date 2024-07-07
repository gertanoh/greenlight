# Use an official Golang runtime as a base image
FROM golang:1.20-alpine AS build


# Set the working directory inside the container
WORKDIR /app

# Install CA certificates on Alpine Linux
# RUN apk --no-cache add ca-certificates
RUN apk update && apk add --no-cache git make ca-certificates


COPY . .

# Build the Golang app using the 'make' command (assuming 'build/api' target is defined in your Makefile)
RUN apk add --no-cache make  # Install 'make' (if not already available in the base image)
RUN make build/api


# Copy the binary from the builder stage to the final stage
FROM scratch
COPY --from=build /app/app .
COPY --from=build /app/.env .


# Expose the port your Golang app listens on
EXPOSE 80


# Use the CMD instruction to set the flag directly and launch your Golang app
CMD ["./app"]
