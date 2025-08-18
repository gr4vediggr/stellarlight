# ---- build ----
FROM node:24-alpine AS build
WORKDIR /app
COPY package*.json ./
RUN npm ci
COPY . .
RUN npm run build

# ---- run (static) ----
FROM nginx:1.27-alpine

# Copy built frontend
COPY --from=build /app/dist /usr/share/nginx/html

# Copy the entrypoint script
COPY docker-entrypoint.sh /docker-entrypoint.sh
RUN chmod +x /docker-entrypoint.sh

# SPA fallback for React Router
RUN printf "server { \
  listen 80; \
  server_name _; \
  root /usr/share/nginx/html; \
  location / { try_files \$uri /index.html; } \
} \
" > /etc/nginx/conf.d/default.conf

# Set the entrypoint
ENTRYPOINT ["/docker-entrypoint.sh"]

EXPOSE 80

# Default command
CMD ["nginx", "-g", "daemon off;"]
