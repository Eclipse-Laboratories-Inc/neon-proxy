FROM golang:1.18 as goproxybuilder

WORKDIR /app

COPY go.mod ./
COPY go.sum ./

RUN go mod download

COPY . ./

RUN make build


FROM ubuntu:22.04 as neon-proxy

RUN apt-get update && apt-get install ca-certificates -y

WORKDIR /app

COPY --from=goproxybuilder /app/bin/neon-proxy /app/

EXPOSE 20501

CMD ["/app/neon-proxy"]


FROM ubuntu:22.04 as neon-subscriber

RUN apt-get update && apt-get install ca-certificates -y

WORKDIR /app

COPY --from=goproxybuilder /app/bin/neon-subscriber /app/

EXPOSE 20503

CMD ["/app/neon-subscriber"]


FROM ubuntu:22.04 as neon-mempool

RUN apt-get update && apt-get install ca-certificates -y

WORKDIR /app

COPY --from=goproxybuilder /app/bin/neon-mempool /app/

EXPOSE 20504

CMD ["/app/neon-mempool"]


FROM ubuntu:22.04 as neon-indexer

RUN apt-get update && apt-get install ca-certificates -y

WORKDIR /app

COPY --from=goproxybuilder /app/bin/neon-indexer /app/

EXPOSE 20502

CMD ["/app/neon-indexer"]


FROM ubuntu:22.04 as neon-wssubscriber

RUN apt-get update && apt-get install ca-certificates -y

WORKDIR /app

COPY --from=goproxybuilder /app/bin/neon-wssubscriber /app/

EXPOSE 20503

CMD ["/app/neon-wssubscriber"]
