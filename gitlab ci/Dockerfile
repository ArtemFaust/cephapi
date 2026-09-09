FROM registry.basealt.ru/alt/alt:p11
LABEL maintainer="Боканёв Артём Владимирович"
RUN apt-get update -y && apt-get upgrade --enable-upgrade -y \
    && apt-get install bash jq yq ceph-base ceph-common librados-devel librbd-devel libcephfs-devel gcc musl-devel -y
RUN useradd -m cephapi -u 1000
#COPY ./linux/aarch64 /home/cephapi/aarch64
#COPY ./linux/amd64 /home/cephapi/amd64
USER 1000
WORKDIR /home/cephapi
CMD ["bash"]