# order_center
#
# VERSION               1.0.1
FROM golang:1.15.5-alpine3.12 as builder1
COPY . /order_center
WORKDIR /order_center
#RUN ls && cd cmd/config && ls

ENV GO111MODULE=on
ENV GOPROXY=https://goproxy.cn,direct
ENV GOPRIVATE=github.com/hzlpypy
ENV GOMOD=/order_center/go.mod

COPY cmd/main.go .
COPY /docker/.ssh /root/.ssh
COPY /docker/.gitconfig /root/.gitconfig
RUN sed -i 's:dl-cdn.alpinelinux.org:mirrors.tuna.tsinghua.edu.cn:g' /etc/apk/repositories && apk add git openssh-client
RUN git config --global --add url."git@github.com:".insteadOf "https://github.com/"
RUN go mod download &&  go build -o ./main .
RUN mkdir /usr/local/order_center && mv main  /usr/local/order_center
WORKDIR /usr/local/order_center
RUN mkdir log
CMD ["./main"]

#
#FROM alpine
#RUN mkdir /usr/local/order_center
#WORKDIR /usr/local/order_center
#RUN mkdir log && chmod 777 log
## --form=0：引用了第一个 stage
#COPY --from=builder1 /order_center/main .
#COPY --from=builder1 /order_center/cmd/config/config.yaml /order_center/cmd/config/config.yaml
##RUN ls && cd /order_center/config && ls
#CMD ["./main"]

#
#FROM alpine
#RUN ls
#RUN mkdir /usr/local/order_center &&  mkdir /usr/local/order_center/log
## --form=0：引用了第一个 stage
#COPY --from=builder1 /order_center/main /usr/local/order_center
#COPY /cmd/config/config.yaml /order_center/cmd/config/config.yaml
#COPY /go  /go
##RUN ls && cd /order_center/config && ls
#CMD ["/usr/local/order_center/main"]