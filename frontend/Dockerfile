FROM node:22-alpine AS deps

WORKDIR /app

COPY frontend/package.json frontend/package-lock.json ./
RUN npm ci

COPY frontend ./

ARG NEXT_PUBLIC_GRAPHQL_API_URL=http://localhost:8000/graphql
ENV NEXT_PUBLIC_GRAPHQL_API_URL=${NEXT_PUBLIC_GRAPHQL_API_URL}

RUN npm run build

FROM node:22-alpine AS runner

WORKDIR /app

ENV NODE_ENV=production
ENV HOSTNAME=0.0.0.0
ENV PORT=3000

COPY --from=deps /app/.next ./.next
COPY --from=deps /app/node_modules ./node_modules
COPY --from=deps /app/package.json ./package.json
COPY --from=deps /app/next.config.ts ./next.config.ts

EXPOSE 3000

CMD ["npm", "run", "start"]
