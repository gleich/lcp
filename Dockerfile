FROM golang:1.22.5

COPY . .

RUN go build && touch .env

CMD ["lcp"]