FROM scratch

ARG SERVICE_PORT

EXPOSE $SERVICE_PORT

COPY short-link-go /

CMD ["/short-link-go"]
