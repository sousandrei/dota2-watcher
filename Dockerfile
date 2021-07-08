FROM golang:alpine as base

FROM base as builder
WORKDIR /opt/dota

ADD .gitignore go.mod go.sum main.go ./

RUN go build
RUN ls
RUN echo 000

FROM base
WORKDIR /opt

COPY --from=builder /opt/dota/dota2-watcher dota

RUN chmod +x /opt/dota

ENTRYPOINT [ "/opt/dota" ]