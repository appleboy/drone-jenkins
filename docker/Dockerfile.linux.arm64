FROM plugins/base:linux-arm64

LABEL maintainer="Bo-Yi Wu <appleboy.tw@gmail.com>" \
  org.label-schema.name="Drone Jenkins" \
  org.label-schema.vendor="Bo-Yi Wu" \
  org.label-schema.schema-version="1.0"

COPY release/linux/arm64/drone-jenkins /bin/

ENTRYPOINT ["/bin/drone-jenkins"]
