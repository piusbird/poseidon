FROM cgr.dev/chainguard/alpine-base:latest

RUN echo -e  "\nhttps://dl-cdn.alpinelinux.org/alpine/edge/community" >> /etc/apk/repositories
RUN apk add --update --no-cache go git make musl-dev curl openssl
RUN mkdir -p /app/src
WORKDIR /app/src
COPY appbuild.sh /app/src
RUN chmod +x appbuild.sh
RUN ./appbuild.sh
COPY intercept.key /app
COPY intercept.csr /app
COPY intercept.crt /app

WORKDIR /
COPY entrypoint.sh /
RUN chmod +x entrypoint.sh
EXPOSE 3000
ENTRYPOINT ./entrypoint.sh
