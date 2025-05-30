#!/bin/bash

echo "" >> .env
echo "PORT=$PORT" >> .env
echo "MONGODB_URI=$MONGODB_URI" >> .env
echo "ENV=$ENV" >> .env
echo "MYSQL_URI=$MYSQL_URI" >> .env
echo "LARAVEL_API_URL=$LARAVEL_API_URL" >> .env
echo "SPACE_DESK_WEBHOOK_X_API_KEY=$SPACE_DESK_WEBHOOK_X_API_KEY" >> .env
echo "SPACE_DESK_API_KEY =$SPACE_DESK_API_KEY " >> .env


echo "[arte arena security] Configurando variáveis de ambiente..."

/app/main
tail -f /dev/null
