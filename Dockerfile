FROM cosmtrek/air:latest

# Install piper TTS for realistic neural text-to-speech
RUN apt-get update && apt-get install -y curl && rm -rf /var/lib/apt/lists/* \
    && curl -L https://github.com/rhasspy/piper/releases/download/2023.11.14-2/piper_linux_x86_64.tar.gz | tar -xzf - -C /opt \
    && ln -s /opt/piper/piper /usr/local/bin/piper

# Download a high-quality English voice model
RUN mkdir -p /opt/piper-voices \
    && curl -L -o /opt/piper-voices/en_US-lessac-medium.onnx https://huggingface.co/rhasspy/piper-voices/resolve/main/en/en_US/lessac/medium/en_US-lessac-medium.onnx \
    && curl -L -o /opt/piper-voices/en_US-lessac-medium.onnx.json https://huggingface.co/rhasspy/piper-voices/resolve/main/en/en_US/lessac/medium/en_US-lessac-medium.onnx.json
WORKDIR /app
ENV air_wd=/app
ENV GOFLAGS="-buildvcs=false"
COPY ./go.mod  ./go.sum /app/
RUN go mod tidy; go mod download
COPY ./src /app/src
CMD ["air"]