FROM alpine:latest

RUN apk add --update --no-cache go git make musl-dev curl openssl
RUN mkdir -p /app/src
WORKDIR /app/src
COPY appbuild.sh /app/src
RUN chmod +x appbuild.sh
RUN ./appbuild.sh
WORKDIR /
COPY entrypoint.sh /
RUN chmod +x entrypoint.sh
EXPOSE 3000
RUN ip addr
ENTRYPOINT ./entrypoint.sh
