FROM nginx:latest

WORKDIR /app

COPY dist .

COPY nginx/conf.d/default.conf.template /etc/nginx/conf.d/default.conf

EXPOSE 8080