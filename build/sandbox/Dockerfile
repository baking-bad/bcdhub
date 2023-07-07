FROM node:16 AS build

ARG TAG
RUN git clone --depth=1 --branch ${TAG} https://github.com/baking-bad/bcd.git /bcd

WORKDIR /bcd
RUN npm i

COPY build/sandbox/env.production /bcd/.env.production
RUN export NODE_OPTIONS=--openssl-legacy-provider
RUN npm run build

FROM nginx:latest AS release
COPY build/sandbox/default.conf /etc/nginx/conf.d/default.conf
COPY build/sandbox/entrypoint.sh /
RUN chmod +x /entrypoint.sh
WORKDIR /usr/share/nginx/html/
COPY --from=build /bcd/dist ./
ENTRYPOINT [ "/entrypoint.sh" ]
