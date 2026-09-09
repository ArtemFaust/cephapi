FROM ubuntu:22.04
LABEL maintainer="Bokanev Artem Vladimirovich"
COPY ./sources.list /etc/apt/sources.list
COPY ./ubuntu-keyring-2012-cdimage.gpg /etc/apt/trusted.gpg.d/ubuntu-keyring-2012-cdimage.gpg
COPY ./ubuntu-keyring-2018-archive.gpg /etc/apt/trusted.gpg.d/ubuntu-keyring-2018-archive.gpg
COPY ./99disable-ssl-check /etc/apt/apt.conf.d/99disable-ssl-check
RUN apt update -y \
    && apt install ca-certificates -y \
    && update-ca-certificates \
    && apt install bash jq yq ceph-base ceph-common librados-dev librbd-dev libcephfs-dev gcc musl-dev -y
RUN useradd -m cephapi -u 1000
USER 1000
WORKDIR /home/cephapi
CMD ["bash"]