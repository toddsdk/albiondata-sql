FROM golang

RUN apt update && apt install -y mysql-client

WORKDIR /go/src/app

COPY . .

RUN go-wrapper download 
RUN go-wrapper install

CMD ["go-wrapper", "run"]

