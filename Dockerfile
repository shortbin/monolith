FROM golang:1.23 AS builder

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .

# CGO_ENABLED=0 to build a static binary (no dynamic linking, dependencies (libc) are included in the binary)
RUN CGO_ENABLED=0 go build -o shortbin ./cmd/api


FROM scratch

COPY --from=builder /app/shortbin .

CMD ["/shortbin"]