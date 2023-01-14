FROM alpine
WORKDIR /
COPY app .
CMD [ "./app" ]
