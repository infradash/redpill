# Dash Dockerfile


# Pull base image.
FROM ubuntu:14.04

# This assumes that the binary has been copied into the directory where this Dockerfile lives.
COPY redpill /usr/bin/

# SSL certs - see https://medium.com/@kelseyhightower/optimizing-docker-images-for-static-binaries-b5696e26eb07
RUN mkdir -p /etc/ssl/certs
COPY ca-certificates.crt /etc/ssl/certs/

# Expose ports.
EXPOSE 5050 5051

# Define mountable directories.
RUN mkdir -p /var/log/redpill/
RUN mkdir -p /var/run/redpill/

VOLUME ["/var/run/redpill", "/var/log/redpill"]

# Define working directory.
WORKDIR /var/run/redpill

# Define default command.
ENTRYPOINT ["redpill", "-logtostderr"]
