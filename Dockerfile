FROM golang:alpine AS build
WORKDIR /opt/src/gocalc
COPY ./main.go ./
RUN go env -w GO111MODULE=auto
# Install git and build-base
RUN go mod init tidy
RUN apk add --no-cache git && apk add --no-cache build-base
# Get repo
RUN go get "github.com/prometheus/client_golang/prometheus" 
RUN go get "github.com/prometheus/client_golang/prometheus/promhttp"
RUN go get "github.com/prometheus/common/log@v0.8.0"
RUN go get "github.com/prometheus/common/version@v0.8.0"
# Test code 
# RUN go test


# Build Go 
RUN go build -o /opt/src/gocalc/app

FROM alpine:latest
COPY --from=build /opt/src/gocalc/app /bin/gocalc
CMD ["/bin/gocalc"]
