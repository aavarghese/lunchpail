FROM docker.io/golang:1.22.6-alpine as builder
WORKDIR /init

COPY go.* .
RUN go mod download

COPY hack/setup/cli.sh .
COPY cmd cmd
COPY pkg pkg
COPY charts charts

RUN ./cli.sh
RUN chmod a+rX lunchpail

FROM docker.io/alpine:3
LABEL org.opencontainers.image.source="https://github.com/IBM/lunchpail"

RUN adduser -u 2000 lunchpail -G root --disabled-password && echo "lunchpail:lunchpail" | chpasswd && chmod -R g=u /home/lunchpail
ENV HOME=/home/lunchpail
WORKDIR /home/lunchpail

# minio server
RUN apk update && apk add --no-cache minio
EXPOSE 9000

COPY --from=builder /init/lunchpail /usr/local/bin/lunchpail

USER lunchpail
CMD ["lunchpail"]