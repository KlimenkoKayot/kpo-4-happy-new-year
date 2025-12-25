#!/bin/bash

# Нагрузочный тест
# Использование: ./test_load.sh [количество_заказов]

BASE_URL="${BASE_URL:-http://localhost:5000}"
ORDER_COUNT=${1:-10}

echo "🚀 Нагрузочный тест Гоzон"
echo "========================="
echo "Количество заказов: $ORDER_COUNT"
echo ""

# Генерируем уникальный User ID для теста
USER_ID=$(uuidgen 2>/dev/null || cat /proc/sys/kernel/random/uuid)
echo "User ID: $USER_ID"

# Создаем аккаунт и пополняем
echo ""
echo "Подготовка: создание аккаунта и пополнение..."
curl -s -X POST "$BASE_URL/api/accounts" -H "X-User-Id: $USER_ID" > /dev/null

DEPOSIT_AMOUNT=$((ORDER_COUNT * 100 + 1000))
curl -s -X POST "$BASE_URL/api/accounts/deposit" \
    -H "X-User-Id: $USER_ID" \
    -H "Content-Type: application/json" \
    -d "{\"amount\": $DEPOSIT_AMOUNT}" > /dev/null

echo "Пополнено на $DEPOSIT_AMOUNT₽"

# Засекаем время
START_TIME=$(date +%s.%N)

# Создаем заказы параллельно
echo ""
echo "Создание $ORDER_COUNT заказов..."

for i in $(seq 1 $ORDER_COUNT); do
    curl -s -X POST "$BASE_URL/api/orders" \
        -H "X-User-Id: $USER_ID" \
        -H "Content-Type: application/json" \
        -d "{\"amount\": 50, \"description\": \"Нагрузочный тест #$i\"}" > /dev/null &

    # Выводим прогресс
    if [ $((i % 10)) -eq 0 ]; then
        echo "  Создано: $i/$ORDER_COUNT"
    fi
done

# Ждем завершения всех запросов
wait

END_TIME=$(date +%s.%N)
DURATION=$(echo "$END_TIME - $START_TIME" | bc)

echo ""
echo "Все заказы созданы за $DURATION секунд"
echo ""

# Ждем обработки
WAIT_TIME=$((ORDER_COUNT / 2 + 5))
echo "Ожидание обработки ($WAIT_TIME сек)..."
sleep $WAIT_TIME

# Проверяем результаты
echo ""
echo "Результаты:"
echo "-----------"

ORDERS_RESPONSE=$(curl -s "$BASE_URL/api/orders" -H "X-User-Id: $USER_ID")
TOTAL_ORDERS=$(echo "$ORDERS_RESPONSE" | grep -o "\"id\"" | wc -l)
FINISHED=$(echo "$ORDERS_RESPONSE" | grep -o "\"status\":\"Finished\"" | wc -l)
CANCELLED=$(echo "$ORDERS_RESPONSE" | grep -o "\"status\":\"Cancelled\"" | wc -l)
NEW=$(echo "$ORDERS_RESPONSE" | grep -o "\"status\":\"New\"" | wc -l)

echo "Всего заказов: $TOTAL_ORDERS"
echo "Оплачено (Finished): $FINISHED"
echo "Отменено (Cancelled): $CANCELLED"
echo "В обработке (New): $NEW"

BALANCE=$(curl -s "$BASE_URL/api/accounts/balance" -H "X-User-Id: $USER_ID" | grep -o "\"balance\":[0-9.]*" | cut -d':' -f2)
echo ""
echo "Остаток на счете: $BALANCE₽"

# Расчет
EXPECTED_SPENT=$((FINISHED * 50))
EXPECTED_BALANCE=$((DEPOSIT_AMOUNT - EXPECTED_SPENT))
echo "Ожидаемый остаток: $EXPECTED_BALANCE₽"

if [ "$BALANCE" == "$EXPECTED_BALANCE" ] || [ "$BALANCE" == "${EXPECTED_BALANCE}.00" ]; then
    echo ""
    echo "✅ Exactly-once семантика работает корректно!"
else
    echo ""
    echo "⚠️  Проверьте расчеты"
fi

echo ""
echo "Производительность: $(echo "scale=2; $ORDER_COUNT / $DURATION" | bc) заказов/сек"
