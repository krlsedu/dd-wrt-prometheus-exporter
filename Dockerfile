FROM golang:1.21-alpine
RUN mkdir /app
ADD . /app
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY *.go ./
RUN CGO_ENABLED=0 GOOS=linux go build -o /dd-wrt-exporter
EXPOSE 2112
CMD ["/dd-wrt-exporter"]
#"-url", "http://192.168.1.1","-username","krlsedu","-password","Su-231018","-interfaces","eth0,eth1,vlan2"
