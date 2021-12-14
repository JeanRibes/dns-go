FROM golang:1.17-alpine as builder

WORKDIR /go/src
RUN mkdir -p progdns


COPY go.mod progdns
COPY go.sum progdns

# téléchargement des dépendances
RUN cd progdns && go mod download

COPY *.go /go/src/progdns/

# build Go
RUN cd progdns && go build -o /main

#FROM alpine
#COPY --from=builder /main .

CMD ["/main", "-addr", "0.0.0.0:53", "-zone", "/zone.db"]