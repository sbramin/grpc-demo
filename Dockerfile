FROM golang:1-alpine AS build

RUN apk update && apk add make git gcc musl-dev

ARG GITHUB_TOKEN
ARG SERVICE
ARG APP

RUN git config --global url."https://${GITHUB_TOKEN}:x-oauth-basic@github.com/".insteadOf "https://github.com/"

ADD . /go/src/github.com/utilitywarehouse/${SERVICE}

WORKDIR /go/src/github.com/utilitywarehouse/${SERVICE}

RUN make clean install
RUN make ${APP}

RUN mv ${APP} /${APP}

FROM alpine:latest

ARG APP

ENV APP=${APP}

RUN apk add --no-cache ca-certificates && mkdir /app
COPY --from=build /${APP} /app/${APP}
#COPY swagger.json /app/swagger.json
ENTRYPOINT /app/${APP}
