FROM golang:1.20-buster AS builder
USER root

# Install jq and git
RUN apt-get update && apt-get install -y jq git && rm -rf /var/lib/apt/lists/*

WORKDIR /tmp/integration_tests

COPY . ./
RUN chmod +x ./set_and_run_tests.sh
CMD ["./set_and_run_tests.sh"]