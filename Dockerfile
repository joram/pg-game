FROM cosmtrek/air:latest
WORKDIR /app
ENV air_wd=/app
ENV GOFLAGS="-buildvcs=false"
COPY ./go.mod  ./go.sum /app/
RUN go mod tidy; go mod download
COPY ./src /app/src
CMD ["air"]