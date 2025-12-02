FROM node:22-alpine

WORKDIR /app

ARG API_URL=http://api-server:8080
ENV API_URL=$API_URL

COPY package.json package-lock.json* ./
RUN npm install

COPY . .

RUN npm run build

EXPOSE 3000

CMD ["npm", "start"]