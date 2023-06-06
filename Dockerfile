FROM  alpine:latest as builder
RUN apk add --update --no-cache curl openssl go && rm -rf /var/cache/apk/*
WORKDIR /app
COPY . ./

RUN go mod download
RUN go build

FROM alpine:latest


RUN apk add --update --no-cache curl openssl && rm -rf /var/cache/apk/*
WORKDIR /app
COPY . ./
RUN rm /app/access.log
COPY --from=builder /app/poseidon /app/poseidon
RUN chmod +x /app/entrypoint.sh

EXPOSE 3000
ENTRYPOINT /app/entrypoint.sh
