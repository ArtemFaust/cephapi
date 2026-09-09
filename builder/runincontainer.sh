#!/bin/bash
if [[ $1 == ""|| $1 == "-h" ]];
then
  echo "./runincontainer.sh run arm64|amd64"
  echo "./runincontainer.sh build arm64|amd64"
  exit 0
fi

# Запуск контейнера
if [[ $1 == "run" ]];
then
  if [[ $2 == "" ]];
  then
    echo "platform ? arm64 or amd64 "
    exit 1
  else
    if [[ $2 == "amd64" ]];
    then
      docker build . --platform=linux/amd64 --tag cephapi_amd64:latest
      docker run --env-file .env --platform linux/$2 --rm -ti  \
      -v ../src/:/opt \
      -v ../dist/:/opt/dist \
      cephapi_amd64:latest bash
    elif [[ $2 == "arm64" ]];
    then
      docker build . --platform=linux/arm64 --tag cephapi_arm64:latest
      docker run  --env-file .env --platform linux/$2 --rm -ti  \
      -v ../src/:/opt \
      -v ../dist/:/opt/dist \
      cephapi_arm64:latest bash
    fi
  fi
fi

# Сборка пакета со статической линковкой
if [[ $1 == "build" ]];
then
  if [[ $2 == "" ]];
  then
    echo "platform ? arm64 or amd64 "
    exit 1
  else
    if [[ $2 == "amd64" ]];
    then
      docker build . --platform=linux/amd64 --tag cephapi_amd64:latest
      docker run --env-file .env --platform linux/$2 --rm -ti  \
      -v ../src/:/opt \
      -v ../dist/:/opt/dist \
      cephapi_amd64:latest
    elif [[ $2 == "arm64" ]];
    then
      docker build . --platform=linux/arm64 --tag cephapi_arm64:latest
      docker run --env-file .env --platform linux/$2 --rm -ti \
       -v ../src/:/opt \
       -v ../dist/:/opt/dist \
       cephapi_arm64:latest
    fi
  fi
fi