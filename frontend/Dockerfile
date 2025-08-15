# ---- build ----
FROM node:24-alpine AS build
WORKDIR /app
COPY package*.json ./
RUN npm ci
COPY . .
RUN npm run build

# ---- run (static) ----
FROM nginx:1.27-alpine
COPY --from=build /app/dist /usr/share/nginx/html
# Single Page App fallback
RUN printf "server { \
  listen 80; \
  server_name _; \
  root /usr/share/nginx/html; \
  location / { try_files \$uri /index.html; } \
} \
" > /etc/nginx/conf.d/default.conf
EXPOSE 80
