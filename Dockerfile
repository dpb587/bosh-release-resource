FROM golang:1.10 as resource
WORKDIR /go/src/github.com/dpb587/bosh-release-resource
COPY . .
ENV CGO_ENABLED=0
RUN mkdir -p /opt/resource
RUN git rev-parse HEAD | tee /opt/resource/version
RUN go build -o /opt/resource/check ./check
RUN go build -o /opt/resource/in ./in
RUN go build -o /opt/resource/out ./out

FROM alpine:3.4 as binaries
RUN apk --no-cache add wget
RUN mkdir /tmp/binaries
RUN true \
  && wget -qO /tmp/binaries/bosh http://s3.amazonaws.com/bosh-cli-artifacts/bosh-cli-3.0.1-linux-amd64 \
  && echo "ccc893bab8b219e9e4a628ed044ebca6c6de9ca0  /tmp/binaries/bosh" | sha1sum -c \
  && chmod +x /tmp/binaries/bosh
RUN true \
  && wget --no-check-certificate -qO /tmp/binaries/jq http://github.com/stedolan/jq/releases/download/jq-1.5/jq-linux64 \
  && echo "d8e36831c3c94bb58be34dd544f44a6c6cb88568  /tmp/binaries/jq" | sha1sum -c \
  && chmod +x /tmp/binaries/jq

FROM alpine:3.4
RUN apk --no-cache add bash ca-certificates git openssh-client
COPY --from=binaries /tmp/binaries /usr/local/bin
COPY --from=resource /opt/resource /opt/resource
ADD tasks/create-dev-release /usr/local/bin
