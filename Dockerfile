FROM golang:1.20 as builder

ARG FLAPPER_VERSION
ENV FLAPPER_VERSION = $FLAPPER_VERSION

WORKDIR /app
COPY go.* ./
RUN go mod download

COPY . ./

RUN CGO_ENABLED=0 go build -mod=readonly -o /flapper

FROM gcr.io/distroless/static-debian11
COPY --from=builder /flapper /
EXPOSE 8080
CMD ["/flapper"]