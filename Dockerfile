FROM alpine:3.2
ADD templates /templates
ADD config-web /config-web
WORKDIR /
ENTRYPOINT [ "/config-web" ]
