FROM plugins/base:windows-amd64

LABEL maintainer="Bo-Yi Wu <appleboy.tw@gmail.com>" \
  org.label-schema.name="Drone Jenkins" \
  org.label-schema.vendor="Bo-Yi Wu" \
  org.label-schema.schema-version="1.0"

COPY release/drone-jenkins.exe /drone-jenkins.exe

ENTRYPOINT [ "\\drone-jenkins.exe" ]
