FROM alpine
COPY Blockchain .
RUN chmod +x Blockchain && apk add --no-cache curl
EXPOSE 80/tcp
CMD [ "./Blockchain" ]