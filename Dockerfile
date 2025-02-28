# build-node image
FROM node:lts-fermium as build-node
COPY public/ /build/public/
COPY Makefile /build/
WORKDIR /build
RUN make fetch-node generate-node

# build-go image
FROM golang:alpine as build-go
RUN apk add --no-cache g++ make
COPY . /build/
COPY --from=build-node /build/public/dist/ /build/public/dist/
WORKDIR /build
RUN make fetch-go generate-go compile

# runtime image
FROM alpine:latest
RUN apk add --no-cache ca-certificates
# set up nsswitch.conf for Go's "netgo" implementation
# - https://github.com/docker-library/golang/blob/1eb096131592bcbc90aa3b97471811c798a93573/1.14/alpine3.12/Dockerfile#L9
#RUN [ ! -e /etc/nsswitch.conf ] && echo 'hosts: files dns' > /etc/nsswitch.conf 
RUN mkdir /app
RUN mkdir /data
COPY --from=build-go /build/flagbrew /app
COPY --from=build-go /build/start.sh /app
COPY --from=build-go /build/.env.required /data/.env

# runtime params
WORKDIR /app
ENV PATH=/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin

EXPOSE 8080

CMD ["/bin/sh", "start.sh"]