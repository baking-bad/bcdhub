FROM node:14 AS build

RUN git clone https://github.com/baking-bad/bcd.git /bcd

WORKDIR /bcd
RUN yarn

COPY build/sandbox/env.production /bcd/.env.production
RUN yarn build

FROM nginx:latest AS release
COPY build/sandbox/default.conf /etc/nginx/conf.d/default.conf
WORKDIR /usr/share/nginx/html/
COPY --from=build /bcd/dist ./