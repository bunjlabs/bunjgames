FROM node:16-alpine3.15 AS builder

RUN apk add --update --no-cache g++ make python2
WORKDIR /app
COPY package.json package-lock.json ./
RUN npm install

ENV NODE_ENV production
ENV PATH="./node_modules/.bin:$PATH"

COPY . ./

RUN npm run build


FROM nginx:alpine

COPY nginx.conf /etc/nginx/conf.d/default.conf
COPY --from=builder /app/public /app

CMD ["nginx", "-g", "daemon off;"]
