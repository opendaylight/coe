FROM alpine
RUN apk add --no-cache python bash openssh-client curl
COPY server.py /server.py
ENTRYPOINT ["python", "server.py"]
