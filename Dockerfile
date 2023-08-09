FROM golang:alpine

MAINTAINER szdyg "szdyg@outlook.com"

ARG GOPROXY=https://goproxy.cn,direct

COPY . /pdb_proxy/

RUN mkdir -p /pdb && \
    cd /pdb_proxy && \
    go build 

WORKDIR /pdb_proxy
ENTRYPOINT ["/pdb_proxy/pdb_proxy"]
