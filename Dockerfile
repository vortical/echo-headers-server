# We statically link all our libraries:
# Build it as:
# CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main .
FROM    scratch
ADD     main /
ENV     PORT 8001
EXPOSE  8001
CMD     ["/main"]
