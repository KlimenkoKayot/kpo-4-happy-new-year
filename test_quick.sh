#!/bin/bash

# Быстрый тест основного сценария
# Использование: ./test_quick.sh

BASE_URL="${BASE_URL:-http://localhost:5000}"
USER_ID="550e8400-e29b-41d4-a716-446655440000"

echo "🛒 Быстрый тест Гоzон"
echo "===================="
echo ""

# 1. Создать аккаунт
echo "1. Создание аккаунта..."
curl -s -X POST "$BASE_URL/api/accounts" \
    -H "X-User-Id: $USER_ID" | head -c 100
echo ""

# 2. Пополнить счет
echo ""
echo "2. Пополнение на 1000₽..."
curl -s -X POST "$BASE_URL/api/accounts/deposit" \
    -H "X-User-Id: $USER_ID" \
    -H "Content-Type: application/json" \
    -d '{"amount": 1000}'
echo ""

# 3. Проверить баланс
echo ""
echo "3. Проверка баланса..."
curl -s "$BASE_URL/api/accounts/balance" \
    -H "X-User-Id: $USER_ID"
echo ""

# 4. Создать заказ
echo ""
echo "4. Создание заказа на 150₽..."
curl -s -X POST "$BASE_URL/api/orders" \
    -H "X-User-Id: $USER_ID" \
    -H "Content-Type: application/json" \
    -d '{"amount": 150, "description": "Тестовый заказ"}'
echo ""

# 5. Ждем обработки
echo ""
echo "5. Ожидание обработки (3 сек)..."
sleep 3

# 6. Проверить заказы
echo ""
echo "6. Список заказов:"
curl -s "$BASE_URL/api/orders" \
    -H "X-User-Id: $USER_ID" | python3 -m json.tool 2>/dev/null || \
curl -s "$BASE_URL/api/orders" -H "X-User-Id: $USER_ID"
echo ""

# 7. Проверить баланс
echo ""
echo "7. Финальный баланс:"
curl -s "$BASE_URL/api/accounts/balance" \
    -H "X-User-Id: $USER_ID"
echo ""
echo ""
echo "✅ Тест завершен!"
