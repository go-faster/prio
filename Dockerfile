FROM ubuntu

COPY priod /usr/local/bin/priod

ENTRYPOINT ["/usr/local/bin/priod"]