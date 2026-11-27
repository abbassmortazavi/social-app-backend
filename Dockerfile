FROM golang:1.27-alpine AS build
WORKDIR /app
COPY . .
RUN go mod tidy && CGO_ENABLED=0 go build -o server .

FROM alpine:3.20
LABEL authors="abbasmortazavi"
WORKDIR /app
COPY --from=build /app/server .
EXPOSE 8085
CMD ["./server"]