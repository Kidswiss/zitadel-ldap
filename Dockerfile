FROM golang:1.25 AS build

WORKDIR /zitadel-ldap

ADD ./ .

RUN make release-glauth-zitadel
RUN GOARCH=amd64 make build_glauth


FROM alpine:3.22

COPY --from=build /zitadel-ldap/bin/zitadel_linux-linux-amd64.so /plugins/zitadel.so
COPY --from=build /zitadel-ldap/gl/glauth /app/glauth
RUN apk add gcompat

ENTRYPOINT ["/app/glauth"]
