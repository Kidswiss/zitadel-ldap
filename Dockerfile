FROM --platform=$BUILDPLATFORM golang:1.25 AS build

WORKDIR /app

COPY go.mod go.sum ./

ARG TARGETOS
ARG TARGETARCH

ENV GOOS=${TARGETOS}
ENV GOARCH=${TARGETARCH}

RUN mkdir -p ./bin
RUN go build -trimpath -tags noembed -ldflags "-s -w" -o ./bin -buildvcs=false github.com/glauth/glauth/v2

COPY . ./

RUN CGO_ENABLED=1 GOOS=${TARGETOS} GOARCH=${TARGETARCH} go build -trimpath -tags noembed -ldflags "-s -w" -buildmode=plugin -o ./bin -buildvcs=false ./lib/...

FROM gcr.io/distroless/base-debian13

COPY --from=build /app/bin/zitadel-ldap.so /plugins/zitadel.so
COPY --from=build /app/bin/glauth /app/glauth

ENTRYPOINT ["/app/glauth"]
CMD ["-c", "/app/config/config.cfg"]
