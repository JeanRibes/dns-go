FROM golang:1.17 as builder

WORKDIR /go/src
RUN mkdir -p progdns


COPY go.mod progdns
COPY go.sum progdns

# téléchargement des dépendances
RUN cd progdns && go mod tidy

COPY *.go /go/src/progdns/

# build Go
RUN CGO_ENABLED=0 GOOS=linux cd progdns && go build -a -installsuffix cgo -o /main

#FROM scratch
#COPY --from=builder /main .

CMD ["/main", "-addr", "0.0.0.0:53", "-zone", "/zone.db"]