FROM alpine:3.2
ADD templates /templates
ADD config-web /trace-web
WORKDIR /
ENTRYPOINT [ "/config-web" ]
