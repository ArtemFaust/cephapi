#!/bin/bash
cd /opt
#CGO_ENABLED=1 CC=gcc GOOS=linux GOARCH=amd64 go build -ldflags="-linkmode external -extldflags '-static'" .
if [[ $(uname -m) == "x86_64" ]];
then
  GOOS=linux GOARCH=amd64 go build -ldflags "-X 'cephapi/utils.SALT=$CEPH_API_SALT'" . 
  mv cephapi dist/linux/amd64/cephapi
elif [[ $(uname -m) == "aarch64" ]];
then
  GOOS=linux GOARCH=arm64 go build -ldflags "-X 'cephapi/utils.SALT=$CEPH_API_SALT'" .
  mv cephapi dist/linux/aarch64/cephapi
fi