FROM nginx:alpine
COPY . .
RUN mv /src/dev/nginx/default.conf /etc/nginx/conf.d/default.conf
RUN mv /src/dev/nginx/nginx.conf /etc/nginx/nginx.conf
