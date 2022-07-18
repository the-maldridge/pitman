FROM docker.io/golang:alpine as build

RUN mkdir -p /go/pitman
COPY ./ /go/pitman
RUN cd /go/pitman && \
        go mod vendor && \
        CGO_ENABLED=0 go build -o /pitman ./

FROM scratch
COPY --from=build /pitman /
COPY theme /theme
COPY forms /forms
ENTRYPOINT ["/pitman"]
EXPOSE 8080/tcp
