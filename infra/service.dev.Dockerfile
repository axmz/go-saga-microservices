FROM golang:1.24.4
ARG SERVICE
WORKDIR /app/services/${SERVICE}
COPY services/${SERVICE}/go.mod services/${SERVICE}/go.sum ./
RUN go install github.com/githubnemo/CompileDaemon@latest
RUN go install github.com/go-delve/delve/cmd/dlv@latest
COPY ./lib /app/lib
COPY ./pkg /app/pkg
RUN go mod download
COPY . /app
EXPOSE 8080 2345
CMD ["CompileDaemon", "--build=go build -o main ./cmd/main.go", "--command=./main", "--directory=.", "--exclude-dir=tmp", "--pattern=\\.go$|\\.html$"] 