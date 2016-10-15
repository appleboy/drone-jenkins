FROM centurylink/ca-certs

ADD drone-jenkins /

ENTRYPOINT ["/drone-jenkins"]
