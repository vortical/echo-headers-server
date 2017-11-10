# We statically link all our libraries:
# Build it as:
# CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main .
#FROM    scratch
FROM alpine:latest
RUN apk add --update curl && rm -rf /var/cache/apk/*
ADD     main /
ENV     PORT 8001
EXPOSE  8001
HEALTHCHECK --interval=10s --timeout=3s --start-period=1s --retries=1 CMD curl -f http://localhost:8001/metrics/health || exit 1
CMD     ["/main"]
