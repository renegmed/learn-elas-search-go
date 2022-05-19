FROM golang:1.18.2-alpine

WORKDIR /go/src/elasticsearch-olivere
COPY . .

RUN go install


CMD [ "elasticsearch-olivere" ]

