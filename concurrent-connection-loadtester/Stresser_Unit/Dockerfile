FROM golang:1.20 as build-env
WORKDIR /app/
COPY . ./
#RUN go mod download
#RUN go get -d -v ./... 
RUN go vet -v ./...
RUN go test -v ./...
RUN CGO_ENABLED=0 go build -o stress_unit main.go
FROM gcr.io/distroless/static
LABEL "microservice.name"="stresser unit"
COPY --from=build-env /app/stress_unit /
COPY --from=build-env /app/config.json /
CMD ["/stress_unit"]