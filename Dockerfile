FROM golang:1.15

ADD src /app
RUN cd /app \
    && go build main.go \
    && mv main /usr/local/bin/generator \
    && chmod +x /usr/local/bin/generator

WORKDIR /app/data