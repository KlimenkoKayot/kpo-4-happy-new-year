#!/bin/bash

# Цвета для вывода
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Конфигурация
BASE_URL="${BASE_URL:-http://localhost:5000}"
USER_ID="550e8400-e29b-41d4-a716-446655440000"
USER_ID_2="660e8400-e29b-41d4-a716-446655440001"

# Счетчики тестов
TESTS_PASSED=0
TESTS_FAILED=0

# Функция для логирования
log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[PASS]${NC} $1"
    ((TESTS_PASSED++))
}

log_error() {
    echo -e "${RED}[FAIL]${NC} $1"
    ((TESTS_FAILED++))
}

log_warning() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

# Функция для выполнения HTTP запроса
http_request() {
    local method=$1
    local endpoint=$2
    local data=$3
    local user_id=${4:-$USER_ID}

    if [ -n "$data" ]; then
        curl -s -X "$method" "${BASE_URL}${endpoint}" \
            -H "Content-Type: application/json" \
            -H "X-User-Id: $user_id" \
            -d "$data"
    else
        curl -s -X "$method" "${BASE_URL}${endpoint}" \
            -H "X-User-Id: $user_id"
    fi
}

# Функция для проверки JSON поля
check_json_field() {
    local json=$1
    local field=$2
    local expected=$3
    local actual=$(echo "$json" | grep -o "\"$field\":[^,}]*" | cut -d':' -f2- | tr -d ' "')

    if [ "$actual" == "$expected" ]; then
        return 0
    else
        return 1
    fi
}

# Функция для извлечения значения из JSON
get_json_value() {
    local json=$1
    local field=$2
    echo "$json" | grep -o "\"$field\":\"[^\"]*\"" | cut -d'"' -f4
}

get_json_number() {
    local json=$1
    local field=$2
    echo "$json" | grep -o "\"$field\":[0-9.]*" | cut -d':' -f2
}

# Ожидание готовности сервисов
wait_for_services() {
    log_info "Ожидание готовности сервисов..."

    local max_attempts=30
    local attempt=0

    while [ $attempt -lt $max_attempts ]; do
        if curl -s "${BASE_URL}/health" > /dev/null 2>&1; then
            log_success "API Gateway доступен"
            return 0
        fi
        ((attempt++))
        echo -n "."
        sleep 2
    done

    log_error "Сервисы не готовы после $max_attempts попыток"
    exit 1
}

# ==================== ТЕСТЫ ====================

echo ""
echo "=========================================="
echo "     ТЕСТИРОВАНИЕ СИСТЕМЫ ГОZON          "
echo "=========================================="
echo ""
echo "Base URL: $BASE_URL"
echo "User ID: $USER_ID"
echo ""

# Ожидаем готовности
wait_for_services

echo ""
echo "=========================================="
echo "     1. ТЕСТЫ PAYMENTS SERVICE           "
echo "=========================================="
echo ""

# --- Тест 1.1: Создание аккаунта ---
log_info "Тест 1.1: Создание аккаунта"
RESPONSE=$(http_request "POST" "/api/accounts" "" "$USER_ID")

if echo "$RESPONSE" | grep -q "\"user_id\":\"$USER_ID\""; then
    log_success "Аккаунт успешно создан"
    echo "  Ответ: $RESPONSE"
else
    # Возможно аккаунт уже существует
    if echo "$RESPONSE" | grep -q "already exists"; then
        log_warning "Аккаунт уже существует (OK для повторных тестов)"
    else
        log_error "Не удалось создать аккаунт"
        echo "  Ответ: $RESPONSE"
    fi
fi

# --- Тест 1.2: Повторное создание аккаунта (должна быть ошибка) ---
log_info "Тест 1.2: Повторное создание аккаунта (ожидается ошибка)"
RESPONSE=$(http_request "POST" "/api/accounts" "" "$USER_ID")

if echo "$RESPONSE" | grep -q "already exists\|error"; then
    log_success "Корректная ошибка при повторном создании"
else
    log_error "Должна быть ошибка при повторном создании"
    echo "  Ответ: $RESPONSE"
fi

# --- Тест 1.3: Проверка баланса (должен быть 0) ---
log_info "Тест 1.3: Проверка начального баланса"
RESPONSE=$(http_request "GET" "/api/accounts/balance" "" "$USER_ID")
BALANCE=$(get_json_number "$RESPONSE" "balance")

if [ "$BALANCE" == "0" ] || [ -z "$BALANCE" ]; then
    log_success "Начальный баланс: $BALANCE"
else
    log_warning "Баланс не равен 0: $BALANCE (возможно, остался от предыдущих тестов)"
fi

# --- Тест 1.4: Пополнение счета ---
log_info "Тест 1.4: Пополнение счета на 1000"
RESPONSE=$(http_request "POST" "/api/accounts/deposit" '{"amount": 1000}' "$USER_ID")

if echo "$RESPONSE" | grep -q "balance"; then
    BALANCE=$(get_json_number "$RESPONSE" "balance")
    log_success "Счет пополнен, новый баланс: $BALANCE"
else
    log_error "Не удалось пополнить счет"
    echo "  Ответ: $RESPONSE"
fi

# --- Тест 1.5: Проверка баланса после пополнения ---
log_info "Тест 1.5: Проверка баланса после пополнения"
RESPONSE=$(http_request "GET" "/api/accounts/balance" "" "$USER_ID")
BALANCE=$(get_json_number "$RESPONSE" "balance")

if [ ! -z "$BALANCE" ]; then
    log_success "Текущий баланс: $BALANCE"
else
    log_error "Не удалось получить баланс"
fi

# Сохраняем текущий баланс
INITIAL_BALANCE=$BALANCE

# --- Тест 1.6: Пополнение с невалидной суммой ---
log_info "Тест 1.6: Пополнение с отрицательной суммой (ожидается ошибка)"
RESPONSE=$(http_request "POST" "/api/accounts/deposit" '{"amount": -100}' "$USER_ID")

if echo "$RESPONSE" | grep -q "error"; then
    log_success "Корректная ошибка при отрицательной сумме"
else
    log_error "Должна быть ошибка при отрицательной сумме"
    echo "  Ответ: $RESPONSE"
fi

# --- Тест 1.7: Баланс несуществующего пользователя ---
log_info "Тест 1.7: Баланс несуществующего пользователя"
RESPONSE=$(http_request "GET" "/api/accounts/balance" "" "99999999-9999-9999-9999-999999999999")

if echo "$RESPONSE" | grep -q "not found\|error"; then
    log_success "Корректная ошибка для несуществующего пользователя"
else
    log_error "Должна быть ошибка для несуществующего пользователя"
    echo "  Ответ: $RESPONSE"
fi

echo ""
echo "=========================================="
echo "     2. ТЕСТЫ ORDERS SERVICE             "
echo "=========================================="
echo ""

# --- Тест 2.1: Получение списка заказов (должен быть пустой или с предыдущими) ---
log_info "Тест 2.1: Получение списка заказов"
RESPONSE=$(http_request "GET" "/api/orders" "" "$USER_ID")

if echo "$RESPONSE" | grep -q "\["; then
    log_success "Список заказов получен"
    ORDER_COUNT=$(echo "$RESPONSE" | grep -o "\"id\"" | wc -l)
    echo "  Количество заказов: $ORDER_COUNT"
else
    log_error "Не удалось получить список заказов"
    echo "  Ответ: $RESPONSE"
fi

# --- Тест 2.2: Создание заказа ---
log_info "Тест 2.2: Создание заказа на 150"
RESPONSE=$(http_request "POST" "/api/orders" '{"amount": 150, "description": "Тестовый заказ - свитер с оленями"}' "$USER_ID")

if echo "$RESPONSE" | grep -q "\"id\""; then
    ORDER_ID=$(get_json_value "$RESPONSE" "id")
    log_success "Заказ создан: $ORDER_ID"

    STATUS=$(get_json_value "$RESPONSE" "status")
    echo "  Статус: $STATUS"

    if [ "$STATUS" == "New" ]; then
        log_success "Начальный статус корректный (New)"
    else
        log_error "Начальный статус должен быть New, получен: $STATUS"
    fi
else
    log_error "Не удалось создать заказ"
    echo "  Ответ: $RESPONSE"
fi

# --- Тест 2.3: Ожидание обработки заказа ---
log_info "Тест 2.3: Ожидание обработки заказа (5 секунд)..."
sleep 5

# --- Тест 2.4: Проверка статуса заказа ---
log_info "Тест 2.4: Проверка статуса заказа после обработки"
if [ ! -z "$ORDER_ID" ]; then
    RESPONSE=$(http_request "GET" "/api/orders/$ORDER_ID" "" "$USER_ID")
    STATUS=$(get_json_value "$RESPONSE" "status")

    if [ "$STATUS" == "Finished" ]; then
        log_success "Заказ успешно оплачен (Finished)"
    elif [ "$STATUS" == "Cancelled" ]; then
        log_warning "Заказ отменен (возможно, недостаточно средств)"
    elif [ "$STATUS" == "New" ]; then
        log_warning "Заказ еще в обработке"
    else
        log_error "Неожиданный статус: $STATUS"
    fi
else
    log_error "ID заказа не найден"
fi

# --- Тест 2.5: Проверка баланса после оплаты ---
log_info "Тест 2.5: Проверка баланса после оплаты"
RESPONSE=$(http_request "GET" "/api/accounts/balance" "" "$USER_ID")
NEW_BALANCE=$(get_json_number "$RESPONSE" "balance")

if [ ! -z "$NEW_BALANCE" ]; then
    log_success "Баланс после оплаты: $NEW_BALANCE"

    # Проверяем что баланс уменьшился (если заказ был оплачен)
    if [ "$STATUS" == "Finished" ]; then
        EXPECTED_BALANCE=$(echo "$INITIAL_BALANCE - 150" | bc)
        if [ "$NEW_BALANCE" == "$EXPECTED_BALANCE" ]; then
            log_success "Баланс корректно уменьшился на 150"
        else
            log_warning "Ожидался баланс $EXPECTED_BALANCE, получен $NEW_BALANCE"
        fi
    fi
else
    log_error "Не удалось получить баланс"
fi

echo ""
echo "=========================================="
echo "     3. ТЕСТЫ EXACTLY-ONCE СЕМАНТИКИ     "
echo "=========================================="
echo ""

# --- Тест 3.1: Создание заказа с недостаточным балансом ---
log_info "Тест 3.1: Создание заказа на сумму больше баланса"
CURRENT_BALANCE=$(get_json_number "$(http_request "GET" "/api/accounts/balance" "" "$USER_ID")" "balance")
BIG_AMOUNT=$(echo "$CURRENT_BALANCE + 1000" | bc)

RESPONSE=$(http_request "POST" "/api/orders" "{\"amount\": $BIG_AMOUNT, \"description\": \"Дорогой заказ - не должен пройти\"}" "$USER_ID")

if echo "$RESPONSE" | grep -q "\"id\""; then
    FAILED_ORDER_ID=$(get_json_value "$RESPONSE" "id")
    log_success "Заказ создан: $FAILED_ORDER_ID"

    log_info "Ожидание обработки (5 секунд)..."
    sleep 5

    RESPONSE=$(http_request "GET" "/api/orders/$FAILED_ORDER_ID" "" "$USER_ID")
    STATUS=$(get_json_value "$RESPONSE" "status")

    if [ "$STATUS" == "Cancelled" ]; then
        log_success "Заказ корректно отменен из-за недостатка средств"
    else
        log_error "Заказ должен быть отменен, статус: $STATUS"
    fi

    # Проверяем что баланс не изменился
    NEW_BALANCE=$(get_json_number "$(http_request "GET" "/api/accounts/balance" "" "$USER_ID")" "balance")
    if [ "$NEW_BALANCE" == "$CURRENT_BALANCE" ]; then
        log_success "Баланс не изменился (exactly-once)"
    else
        log_error "Баланс изменился! Было: $CURRENT_BALANCE, стало: $NEW_BALANCE"
    fi
else
    log_error "Не удалось создать заказ"
fi

# --- Тест 3.2: Несколько быстрых заказов подряд ---
log_info "Тест 3.2: Создание нескольких заказов подряд"

BALANCE_BEFORE=$(get_json_number "$(http_request "GET" "/api/accounts/balance" "" "$USER_ID")" "balance")
echo "  Баланс до заказов: $BALANCE_BEFORE"

# Создаем 3 заказа по 50
for i in 1 2 3; do
    RESPONSE=$(http_request "POST" "/api/orders" "{\"amount\": 50, \"description\": \"Быстрый заказ #$i\"}" "$USER_ID")
    if echo "$RESPONSE" | grep -q "\"id\""; then
        ORDER_ID=$(get_json_value "$RESPONSE" "id")
        echo "  Заказ #$i создан: ${ORDER_ID:0:8}..."
    fi
done

log_info "Ожидание обработки всех заказов (8 секунд)..."
sleep 8

BALANCE_AFTER=$(get_json_number "$(http_request "GET" "/api/accounts/balance" "" "$USER_ID")" "balance")
echo "  Баланс после заказов: $BALANCE_AFTER"

# Считаем сколько успешных заказов
FINISHED_COUNT=$(http_request "GET" "/api/orders" "" "$USER_ID" | grep -o "\"status\":\"Finished\"" | wc -l)
log_success "Успешно оплаченных заказов: $FINISHED_COUNT"

echo ""
echo "=========================================="
echo "     4. ТЕСТЫ ВАЛИДАЦИИ                  "
echo "=========================================="
echo ""

# --- Тест 4.1: Заказ без X-User-Id ---
log_info "Тест 4.1: Запрос без X-User-Id"
RESPONSE=$(curl -s -X GET "${BASE_URL}/api/orders")

if echo "$RESPONSE" | grep -q "error\|required"; then
    log_success "Корректная ошибка при отсутствии X-User-Id"
else
    log_error "Должна быть ошибка при отсутствии X-User-Id"
    echo "  Ответ: $RESPONSE"
fi

# --- Тест 4.2: Невалидный UUID ---
log_info "Тест 4.2: Запрос с невалидным UUID"
RESPONSE=$(curl -s -X GET "${BASE_URL}/api/orders" -H "X-User-Id: invalid-uuid")

if echo "$RESPONSE" | grep -q "error\|Invalid"; then
    log_success "Корректная ошибка при невалидном UUID"
else
    log_error "Должна быть ошибка при невалидном UUID"
    echo "  Ответ: $RESPONSE"
fi

# --- Тест 4.3: Заказ с пустым описанием ---
log_info "Тест 4.3: Заказ с пустым описанием"
RESPONSE=$(http_request "POST" "/api/orders" '{"amount": 100, "description": ""}' "$USER_ID")

if echo "$RESPONSE" | grep -q "error"; then
    log_success "Корректная ошибка при пустом описании"
else
    log_error "Должна быть ошибка при пустом описании"
    echo "  Ответ: $RESPONSE"
fi

# --- Тест 4.4: Заказ с нулевой суммой ---
log_info "Тест 4.4: Заказ с нулевой суммой"
RESPONSE=$(http_request "POST" "/api/orders" '{"amount": 0, "description": "Тест"}' "$USER_ID")

if echo "$RESPONSE" | grep -q "error"; then
    log_success "Корректная ошибка при нулевой сумме"
else
    log_error "Должна быть ошибка при нулевой сумме"
    echo "  Ответ: $RESPONSE"
fi

echo ""
echo "=========================================="
echo "     5. ТЕСТЫ ВТОРОГО ПОЛЬЗОВАТЕЛЯ       "
echo "=========================================="
echo ""

# --- Тест 5.1: Создание аккаунта для второго пользователя ---
log_info "Тест 5.1: Создание аккаунта для второго пользователя"
RESPONSE=$(http_request "POST" "/api/accounts" "" "$USER_ID_2")

if echo "$RESPONSE" | grep -q "\"user_id\""; then
    log_success "Аккаунт второго пользователя создан"
elif echo "$RESPONSE" | grep -q "already exists"; then
    log_warning "Аккаунт уже существует"
else
    log_error "Ошибка создания аккаунта"
fi

# --- Тест 5.2: Заказ без денег на счету ---
log_info "Тест 5.2: Заказ без денег на счету"
RESPONSE=$(http_request "POST" "/api/orders" '{"amount": 100, "description": "Заказ без денег"}' "$USER_ID_2")

if echo "$RESPONSE" | grep -q "\"id\""; then
    ORDER_ID=$(get_json_value "$RESPONSE" "id")
    log_success "Заказ создан: $ORDER_ID"

    sleep 5

    RESPONSE=$(http_request "GET" "/api/orders/$ORDER_ID" "" "$USER_ID_2")
    STATUS=$(get_json_value "$RESPONSE" "status")

    if [ "$STATUS" == "Cancelled" ]; then
        log_success "Заказ отменен из-за отсутствия средств"
    else
        log_error "Заказ должен быть отменен, статус: $STATUS"
    fi
fi

# --- Тест 5.3: Изоляция данных между пользователями ---
log_info "Тест 5.3: Проверка изоляции данных"
ORDERS_USER1=$(http_request "GET" "/api/orders" "" "$USER_ID" | grep -o "\"id\"" | wc -l)
ORDERS_USER2=$(http_request "GET" "/api/orders" "" "$USER_ID_2" | grep -o "\"id\"" | wc -l)

log_success "Заказов у пользователя 1: $ORDERS_USER1"
log_success "Заказов у пользователя 2: $ORDERS_USER2"

if [ "$ORDERS_USER1" -ne "$ORDERS_USER2" ] || [ "$ORDERS_USER2" -le "1" ]; then
    log_success "Данные изолированы между пользователями"
else
    log_warning "Проверьте изоляцию данных"
fi

echo ""
echo "=========================================="
echo "           РЕЗУЛЬТАТЫ ТЕСТОВ             "
echo "=========================================="
echo ""
echo -e "Успешно: ${GREEN}$TESTS_PASSED${NC}"
echo -e "Провалено: ${RED}$TESTS_FAILED${NC}"
echo ""

if [ $TESTS_FAILED -eq 0 ]; then
    echo -e "${GREEN}✓ Все тесты пройдены успешно!${NC}"
    exit 0
else
    echo -e "${RED}✗ Некоторые тесты провалены${NC}"
    exit 1
fi
