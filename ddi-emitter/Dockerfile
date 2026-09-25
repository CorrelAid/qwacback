FROM node:22-alpine

WORKDIR /app

# Copy only the manifest first so dependency installation is cached across
# source-only changes.
COPY package.json package-lock.json ./
RUN npm ci --omit=dev --no-audit --no-fund

COPY server.mjs ./

ENV DDI_EMITTER_PORT=8091
EXPOSE 8091

USER node
CMD ["node", "server.mjs"]
