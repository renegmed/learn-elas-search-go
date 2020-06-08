FROM golang:1.12.5-alpine

WORKDIR /go/src/elasticsearch-olivere
COPY . .

RUN go install


CMD [ "elasticsearch-olivere" ]

