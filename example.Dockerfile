FROM ubuntu

COPY prio-example /usr/local/bin/prio-example

ENTRYPOINT ["/usr/local/bin/prio-example"]