FROM golang:latest AS builder

RUN mkdir /app
WORKDIR /app
COPY . .
RUN go get -v github.com/benmatselby/go-azuredevops/azuredevops && \
    go get -v github.com/llgcode/draw2d && \
    go get -v github.com/Azure/azure-storage-blob-go/azblob && \
    CGO_ENABLED=0 GOOS=linux go build -a -o devops .

FROM scratch
COPY --from=builder /app/devops .
COPY --from=builder /app/luxisr.ttf .

CMD ["/devops", "-h"]