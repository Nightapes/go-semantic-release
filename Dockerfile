FROM alpine:3.14

WORKDIR /code

COPY ./build/go-semantic-release.linux_x86_64 /usr/local/bin/go-semantic-release

USER 1000

ENTRYPOINT [ "go-semantic-release" ]
