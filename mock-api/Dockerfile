FROM golang:1.23  AS build-stage

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY *.go ./
COPY token ./token
COPY stage ./stage

RUN CGO_ENABLED=0 GOOS=linux go build -o /ase-prep

FROM gcr.io/distroless/base-debian12 AS build-release-stage

WORKDIR / 

COPY --from=build-stage /ase-prep /ase-prep

ENTRYPOINT  ["/ase-prep"]
