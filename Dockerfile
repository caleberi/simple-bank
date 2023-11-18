#[build stage]
# select base image 
FROM golang:1.21-alpine3.18 AS builder

# create workdir
WORKDIR /app

# copy application to the workdir

COPY . .

# build executable binary file 

RUN go build -o simple_bank main.go


#[run stage]

FROM alpine:3.18

# create workdir
WORKDIR /app

COPY --from=builder /app/simple_bank .

# copy env
COPY dev.env .

# expose port address

EXPOSE 8080

# execute binary

CMD [ "/app/simple_bank" ]

