# Build the policy service and package it with its Rego packs.
# Public/community base images only; no authenticated Red Hat registry.
FROM docker.io/library/golang:1.26 AS build
WORKDIR /src
ENV GOFLAGS=-mod=mod GOTOOLCHAIN=local CGO_ENABLED=0
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -trimpath -ldflags="-s -w" -o /out/policyd ./cmd/policyd

FROM gcr.io/distroless/static-debian12:nonroot
COPY --from=build /out/policyd /usr/local/bin/policyd
COPY --from=build /src/policies /policies
ENV POLICY_ADDR=:8181 POLICY_DIR=/policies
EXPOSE 8181
USER nonroot:nonroot
ENTRYPOINT ["/usr/local/bin/policyd"]
