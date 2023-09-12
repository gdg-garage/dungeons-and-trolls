FROM --platform=linux/amd64 ubuntu:latest

RUN apt update && apt install -y \
golang \
ca-certificates \
&& \
apt-get clean && \ 
rm -rf /var/lib/apt/lists/*

WORKDIR /app

COPY . .
RUN go mod download && go mod verify

RUN go build -v -o dnt .

ENTRYPOINT [ "./dnt" ]
