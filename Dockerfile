FROM golang:1.24 AS build

ENV CGO_ENABLED=1

WORKDIR /zitadel-ldap

ADD ./ .

RUN make release-glauth-zitadel
RUN CGO_ENABLED=0 GOARCH=amd64 make build_glauth


FROM alpine:3.21

COPY --from=build /zitadel-ldap/bin/zitadel_linux-linux-amd64.so /plugins/zitadel.so
COPY --from=build /zitadel-ldap/gl/glauth /app/glauth

ENTRYPOINT ["/app/glauth"]
