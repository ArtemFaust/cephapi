FROM golang:latest
COPY entrypoint.sh ./entrypoint.sh
RUN apt update -y && apt upgrade -y \
    && apt install ceph-base ceph-common librados-dev librbd-dev libcephfs-dev gcc musl-dev -y \
    && chmod +x entrypoint.sh
CMD ["./entrypoint.sh"]