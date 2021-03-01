FROM buildpack-deps:scm as snow

RUN ["git", "clone", "https://github.com/scottschiller/Snowstorm.git"]
RUN ["git", "-C", "Snowstorm", "checkout", "862ff2939eb05feb219bd47c7cb760237038bbad"]


FROM node:alpine as deps

RUN ["mkdir", "/deps"]
COPY package.json package-lock.json /deps/
WORKDIR /deps
RUN ["npm", "install"]


FROM alpine as index

RUN ["apk", "add", "perl"]
COPY index.html /
RUN ["perl", "-p0i", "-e", "s/<!--\\s*-->//g", "index.html"]
RUN ["perl", "-pi", "-e", "s/^\\s+//", "index.html"]


FROM nginx:alpine

COPY --from=deps /deps/node_modules/bootstrap/dist/css/bootstrap.min.css /usr/share/nginx/html/
COPY --from=snow /Snowstorm/snowstorm-min.js /usr/share/nginx/html/
COPY --from=index /index.html /usr/share/nginx/html/

COPY nginx.conf /etc/nginx/conf.d/
COPY style.css favicon.jpg bg.jpg script.js robots.txt /usr/share/nginx/html/
