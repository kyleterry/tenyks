FROM golang:1.7
MAINTAINER Kyle Terry "kyle@kyleterry.com"
COPY . /go/src/github.com/kyleterry/tenyks
WORKDIR /go/src/github.com/kyleterry/tenyks
RUN apt-get update -yqq && apt-get install -y libzmq3-dev pkg-config
RUN make clean && make && make install
EXPOSE 61123 61124 12666
ENTRYPOINT ["tenyks"]
