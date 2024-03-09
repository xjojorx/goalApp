FROM golang:1.22

WORKDIR /app

COPY . .

RUN go mod download

RUN GOOS=linux go build -o /goalApp

EXPOSE 42069

CMD ["/goalApp"]
