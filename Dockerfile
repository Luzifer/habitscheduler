FROM golang

MAINTAINER Knut Ahlers <knut@ahlers.me>

RUN go get -v github.com/Luzifer/habitscheduler && \
    go install github.com/Luzifer/habitscheduler

EXPOSE 3000
ENTRYPOINT ["/go/bin/habitscheduler"]
CMD ["--"]
