#### Prepare stage
FROM golang:buster as base-build
RUN apt-get update && \
    apt-get install -y g++ make git curl

FROM chromedp/headless-shell:latest as base-release
RUN apt-get update && apt-get install -y tini ca-certificates curl

#### Build stage
FROM base-build as builder
ARG BRANCH
ARG COMMIT
ENV BRANCH=$BRANCH \
    COMMIT=$COMMIT
WORKDIR /app
COPY . .
RUN go mod download
RUN echo "Build for Linux"; go build -v -o sample main.go

#### Release stage
FROM base-release
WORKDIR /app
COPY --from=builder /app/sample /app/
ENTRYPOINT ["tini", "--"]
CMD ["/app/sample"]