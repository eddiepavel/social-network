# ---- Stage 1: Dependencies ----
FROM node:20-alpine AS deps

WORKDIR /app

COPY package.json package-lock.json ./
RUN npm ci

# ---- Stage 2: Build ----
FROM node:20-alpine AS builder

WORKDIR /app

# Accept API base URL as build arg (defaults to localhost for local dev)
ARG NEXT_PUBLIC_API_BASE=http://localhost:8000
ENV NEXT_PUBLIC_API_BASE=${NEXT_PUBLIC_API_BASE}

# Bring in installed node_modules
COPY --from=deps /app/node_modules ./node_modules
COPY . .

# Create .env from .env.example, then override with build arg
RUN cp .env.example .env \
    && sed -i "s|^NEXT_PUBLIC_API_BASE=.*|NEXT_PUBLIC_API_BASE=${NEXT_PUBLIC_API_BASE}|" .env

# Enable standalone output for a minimal production image
RUN sed -i 's/const nextConfig = {/const nextConfig = {\n  output: "standalone",/' next.config.mjs

RUN npm run build

# ---- Stage 3: Runtime ----
FROM node:20-alpine AS runner

WORKDIR /app

ENV NODE_ENV=production

# Don't run as root
RUN addgroup --system --gid 1001 nodejs \
    && adduser --system --uid 1001 nextjs

# Copy the standalone server output
COPY --from=builder /app/.next/standalone ./

# Copy static assets and public files
COPY --from=builder /app/.next/static ./.next/static
COPY --from=builder /app/public ./public

USER nextjs

EXPOSE 3000

ENV HOSTNAME="0.0.0.0"

CMD ["node", "server.js"]
