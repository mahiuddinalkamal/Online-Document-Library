FROM alpine:latest

RUN apk update

RUN apk add --no-cache nginx

RUN mkdir -p /run/nginx

COPY nginx.conf /etc/nginx/

EXPOSE 80

CMD ["nginx"]
