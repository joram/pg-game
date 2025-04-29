FROM cosmtrek/air:latest
WORKDIR /app
ENV air_wd=/app
ENV GOFLAGS="-buildvcs=false"
COPY .air.toml /app/.air.toml
COPY ./go.mod  ./go.sum /app/
RUN go mod tidy; go mod download
CMD ["air"]