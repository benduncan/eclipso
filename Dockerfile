FROM golang:1.17-alpine AS build-env

# Build phase
RUN apk add build-base git

ENV DNS_PORT 53/udp

ADD ./ /workspace/eclipso
WORKDIR /workspace/eclipso

#RUN make clean
RUN make build

RUN apk del build-base git

# Next, just copy the golang binary, create a lightweight environment

FROM alpine
WORKDIR /workspace/eclipso
RUN apk add ca-certificates
COPY --from=build-env /workspace/eclipso/bin/ /workspace/eclipso/bin/

EXPOSE $API_PORT
ENTRYPOINT ["/workspace/eclipso/bin/eclipso"]