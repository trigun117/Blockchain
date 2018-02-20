FROM alpine
COPY Blockchain .
RUN chmod +x Blockchain
EXPOSE 80/tcp
CMD [ "./Blockchain" ]