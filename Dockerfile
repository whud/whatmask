FROM golang:1.24-alpine AS build
WORKDIR /src
COPY go.mod ./
COPY cmd/ cmd/
COPY internal/ internal/
RUN CGO_ENABLED=0 go build -ldflags="-s -w" -o /whatmask ./cmd/whatmask/

FROM scratch
COPY --from=build /whatmask /whatmask
EXPOSE 8080
ENTRYPOINT ["/whatmask", "--serve"]
