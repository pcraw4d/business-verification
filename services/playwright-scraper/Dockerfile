FROM mcr.microsoft.com/playwright:v1.40.0-focal

WORKDIR /app

# Copy package files
COPY package*.json ./

# Install dependencies (don't use --production to ensure playwright is available)
RUN npm install

# Ensure Playwright browsers are installed (base image should have them, but verify)
RUN npx playwright install chromium || true

# Copy application
COPY . .

# Expose port
EXPOSE 3000

# Start service
CMD ["npm", "start"]

