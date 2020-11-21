# build stage
FROM golang:alpine AS build-env
RUN apk --no-cache add git
ADD . /src
RUN cd /src && go build -o dms ./cmd/dms.go

# final stage
FROM alpine
WORKDIR /app
COPY --from=build-env /src/dms /app/
ENTRYPOINT ./dms