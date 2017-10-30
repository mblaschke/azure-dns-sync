#############################
# Build container (multi stage build)
# will only be used for building
# the application and is not part
# of the final image
#############################
FROM golang as builder
RUN go get github.com/Azure/azure-sdk-for-go/arm/dns
RUN go get github.com/bogdanovich/dns_resolver
RUN go get gopkg.in/yaml.v2
RUN go get github.com/jessevdk/go-flags
RUN go get github.com/robfig/cron
RUN mkdir /go/go-azure-dns-sync
COPY ./ /go/go-azure-dns-sync
WORKDIR /go/go-azure-dns-sync
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main .


#############################
# Final image
#############################
FROM alpine
COPY --from=builder /go/go-azure-dns-sync/main /usr/bin/azure-dns-sync
CMD ["azure-dns-sync"]
