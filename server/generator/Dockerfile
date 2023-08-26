FROM --platform=linux/amd64 ubuntu:latest

RUN apt update && apt install -y \
libdeflate0 \
libjbig0 \
&& \
apt-get clean && \ 
rm -rf /var/lib/apt/lists/*

COPY dntgenerator /
COPY libcage-core.so /

ENTRYPOINT [ "/dntgenerator" ]