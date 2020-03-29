FROM ubuntu:18.04

MAINTAINER dronbrigadir

RUN apt-get -y update
RUN apt install -y git wget gcc gnupg

RUN wget https://dl.google.com/go/go1.13.8.linux-amd64.tar.gz
RUN tar -xvf go1.13.8.linux-amd64.tar.gz
RUN mv go /usr/local

ENV GOROOT /usr/local/go
ENV GOPATH $HOME/go
ENV PATH $GOPATH/bin:$GOROOT/bin:$PATH

WORKDIR /server
COPY . .

EXPOSE 5000

RUN go build ./cmd/main.go

CMD ./main