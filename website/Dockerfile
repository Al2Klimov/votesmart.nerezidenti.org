FROM node:alpine as deps

RUN ["mkdir", "/deps"]
COPY package.json package-lock.json /deps/
WORKDIR /deps
RUN ["npm", "install"]


FROM nginx:alpine

COPY --from=deps /deps/node_modules/bootstrap/dist/css/bootstrap.min.css /usr/share/nginx/html/
COPY nginx.conf /etc/nginx/conf.d/
COPY index.html style.css favicon.jpg bg.jpg robots.txt /usr/share/nginx/html/
