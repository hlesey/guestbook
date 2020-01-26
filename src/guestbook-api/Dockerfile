FROM golang:latest as builder 

WORKDIR /go/src/guestbook-api
COPY . .
RUN go build 


FROM golang:latest 

WORKDIR /app
COPY --from=builder /go/src/guestbook-api/guestbook-api /app
RUN chmod 500 /app/guestbook-api

EXPOSE 8080
CMD [ "/app/guestbook-api" ]