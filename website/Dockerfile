FROM buildpack-deps:scm as snow

RUN ["git", "clone", "https://github.com/scottschiller/Snowstorm.git"]
RUN ["git", "-C", "Snowstorm", "checkout", "862ff2939eb05feb219bd47c7cb760237038bbad"]


FROM node:alpine as deps

RUN ["mkdir", "/deps"]
COPY package.json package-lock.json /deps/
WORKDIR /deps
RUN ["npm", "install"]


FROM nginx:alpine

COPY --from=deps /deps/node_modules/bootstrap/dist/css/bootstrap.min.css /usr/share/nginx/html/
COPY --from=snow /Snowstorm/snowstorm-min.js /usr/share/nginx/html/

COPY nginx.conf /etc/nginx/conf.d/
COPY index.html style.css favicon.jpg bg.jpg script.js robots.txt /usr/share/nginx/html/
