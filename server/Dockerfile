FROM --platform=linux/amd64 golang:latest

RUN apt update && apt install -y \
libdeflate0 \
libjbig0 \
&& \
apt-get clean && \ 
rm -rf /var/lib/apt/lists/*

WORKDIR /app

# COPY . .
# RUN go mod download && go mod verify

# RUN go build -v -o dnt .

# ENTRYPOINT [ "./dnt" ]