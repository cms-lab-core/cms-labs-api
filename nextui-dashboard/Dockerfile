FROM docker.io/library/nginx:1.31.6-alpine3.24

RUN apk del --no-cache \
      nginx-module-acme \
      nginx-module-geoip \
      nginx-module-image-filter \
      nginx-module-njs \
      nginx-module-xslt \
    && apk upgrade --no-cache

COPY dist /usr/share/nginx/html/
COPY nginx /etc/nginx/
COPY 20-envsubst-on-templates.sh /docker-entrypoint.d

# Открываем порт 80
EXPOSE 80
EXPOSE 443
