FROM node:20.11.1-alpine3.19 AS builder

WORKDIR /app


COPY package*.json ./


RUN npm ci


COPY . .

ARG VITE_API_BASE_URL
ENV VITE_API_BASE_URL=$VITE_API_BASE_URL

RUN npm run build


FROM nginx:1.25.4-alpine3.18


COPY --from=builder /app/dist /usr/share/nginx/html

EXPOSE 80

CMD ["nginx", "-g", "daemon off;"]