# ==============================================================================
# ========= build golang project ===============================================
# ==============================================================================
FROM golang:latest as buildgo
WORKDIR /app

# ========= build project ======================================================

# install dependencies
COPY go.mod go.sum ./
RUN go mod download

# copy and build source
COPY ./cache/ ./cache/
COPY ./cli/  ./cli/
COPY ./config/ ./config/
COPY ./handlers/ ./handlers/
COPY ./models/ ./models/
COPY ./server/ ./server/
COPY ./utils/ ./utils/
COPY ./main.go .

RUN go build -o openefs .

# ========= test project =======================================================

RUN go test ./...



# ==============================================================================
# ========= test python code ===================================================
# ==============================================================================
FROM tensorflow/tensorflow:latest-py3 as testpython
WORKDIR /app

# TODO: write tests



# ==============================================================================
# ========= build execution-environment ========================================
# ==============================================================================
FROM alpine:latest as runner
RUN apk --no-cache add ca-certificates
WORKDIR /app
COPY --from=buildgo /app/openefs .
# TODO: copy relevant python code

CMD ["./openefs"]  