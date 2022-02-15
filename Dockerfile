FROM golang:1.17-alpine

WORKDIR /app
COPY go.mod ./
RUN go mod download
COPY *.go ./
RUN go build -o /httpprx
EXPOSE 6060
CMD [ "/httpprx" ]