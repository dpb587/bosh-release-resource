FROM golang:1.10 as resource
WORKDIR /go/src/github.com/dpb587/bosh-release-resource
COPY . .
ENV CGO_ENABLED=0
RUN mkdir -p /opt/resource
RUN git rev-parse HEAD | tee /opt/resource/version
RUN go build -o /opt/resource/check check/*.go
RUN go build -o /opt/resource/in in/*.go
RUN go build -o /opt/resource/out out/*.go

FROM alpine:3.4 as binaries
RUN apk --no-cache add wget
RUN true \
  && wget -qO /usr/local/bin/bosh http://s3.amazonaws.com/bosh-cli-artifacts/bosh-cli-3.0.1-linux-amd64 \
  && echo "ccc893bab8b219e9e4a628ed044ebca6c6de9ca0  /usr/local/bin/bosh" | sha1sum -c \
  && chmod +x /usr/local/bin/bosh

FROM alpine:3.4
RUN apk --no-cache add ca-certificates git openssh-client
COPY --from=binaries /usr/local/bin/bosh /usr/local/bin
COPY --from=resource /opt/resource /opt/resource
