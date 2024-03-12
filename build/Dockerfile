ARG RUN_FROM=gcr.io/distroless/static-debian11@sha256:8ad6f3ec70dad966479b9fb48da991138c72ba969859098ec689d1450c2e6c97
FROM ${RUN_FROM} as image

COPY output/bin/digicert-ca-connector /bin
COPY manifest.json /bin

USER nonroot:nonroot

ENTRYPOINT [ "/bin/digicert-ca-connector" ]
