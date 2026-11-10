FROM oven/bun:1.4.4 AS client-build
WORKDIR /client
COPY apps/client/package.json apps/client/bun.lock ./
RUN bun install --frozen-lockfile
COPY apps/client/ .
RUN bun run build

FROM golang:1.26.10-alpine AS api-build
ARG TARGETOS=linux
ARG TARGETARCH
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH:-amd64} \
    go build -trimpath -ldflags="-s -w -X github.com/FacileStudio/sonar/cmd.version=$(sh scripts/version.sh CHANGELOG.md)" \
    -o /out/sonar .

FROM gcr.io/distroless/static-debian12:nonroot
COPY --from=api-build /out/sonar /sonar
COPY --from=client-build /client/build /client
ENV CLIENT_DIR=/client PORT=4000
EXPOSE 4000
USER nonroot:nonroot
ENTRYPOINT ["/sonar"]
CMD ["serve"]
