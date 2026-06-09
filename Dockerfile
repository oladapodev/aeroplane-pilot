FROM golang:1.25-bookworm AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o /out/pilot ./cmd/pilot

FROM gcr.io/distroless/static-debian12
WORKDIR /
COPY --from=build /out/pilot /pilot
ENTRYPOINT ["/pilot"]
