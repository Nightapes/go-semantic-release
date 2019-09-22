FROM alpine:3.10.2

WORKDIR /code

COPY ./build/go-semantic-release .

USER 1000

ENTRYPOINT [ "./go-semantic-release" ]