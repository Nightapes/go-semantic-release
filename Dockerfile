FROM alpine:3.10.2

WORKDIR /code

COPY ./build/go-semantic-release.linux_x86_64 .

USER 1000

ENTRYPOINT [ "./go-semantic-release.linux_x86_64" ]