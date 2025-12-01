FROM alpine:3.22

ARG TARGETOS
ARG TARGETARCH

LABEL maintainer="Bo-Yi Wu <appleboy.tw@gmail.com>" \
  org.label-schema.name="Drone Jenkins Plugin" \
  org.label-schema.vendor="Bo-Yi Wu" \
  org.label-schema.schema-version="1.0"

LABEL org.opencontainers.image.source=https://github.com/appleboy/drone-jenkins
LABEL org.opencontainers.image.description="Drone Jenkins"
LABEL org.opencontainers.image.licenses=MIT

RUN apk add --no-cache ca-certificates && \
  rm -rf /var/cache/apk/*

RUN addgroup -g 1000 drone && \
  adduser -D -u 1000 -G drone drone

COPY release/${TARGETOS}/${TARGETARCH}/drone-jenkins /bin/
RUN chown drone:drone /bin/drone-jenkins

USER drone

ENTRYPOINT ["/bin/drone-jenkins"]
