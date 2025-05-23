FROM node:20-alpine

WORKDIR /app

# Установка инструментов для возможного обращения к системным библиотекам
RUN apk add --no-cache git make curl

# Конфигурация npm
ENV PATH /app/node_modules/.bin:$PATH

# Добавляем поддержку extra_hosts для доступа к хост-машине
# (в docker-compose.yml используется network_mode: "host" для этой же цели)

# Порт для разработки
EXPOSE 5173

# Настройки по умолчанию
ENV NODE_ENV=development
ENV VITE_HOST=0.0.0.0
ENV VITE_ENABLE_CORS_DEBUG=true
ENV DOCKER_ENV=true

# Копирование package.json и package-lock.json
COPY package*.json ./

# Установка зависимостей
RUN npm ci

# Команда по умолчанию - добавлено --strictPort=false для автоматического выбора порта если 5173 занят
CMD ["npm", "run", "dev", "--", "--host", "0.0.0.0", "--strictPort=false"] 