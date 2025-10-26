#!/bin/bash

# ============================================
# CURL ТЕСТЫ ДЛЯ GOPHERMART
# ============================================

BASE_URL="http://localhost:8080"

echo "🚀 Gophermart API Tests"
echo "======================="
echo ""

# ============================================
# 1. РЕГИСТРАЦИЯ
# ============================================

echo "📝 1. Регистрация пользователя"
echo "================================"

REGISTER_RESPONSE=$(curl -s -i -X POST "$BASE_URL/api/user/register" \
  -H "Content-Type: application/json" \
  -d '{
    "login": "testuser",
    "password": "password123"
  }')

echo "$REGISTER_RESPONSE"
echo ""

# Извлекаем токен из заголовка Authorization
TOKEN=$(echo "$REGISTER_RESPONSE" | grep -i "Authorization:" | awk '{print $3}' | tr -d '\r')

if [ -z "$TOKEN" ]; then
    echo "❌ Не удалось получить токен!"
    exit 1
fi

echo "✅ Токен получен: $TOKEN"
echo ""

# ============================================
# 2. ПОПЫТКА ПОВТОРНОЙ РЕГИСТРАЦИИ (409)
# ============================================

echo "📝 2. Попытка повторной регистрации (должен вернуть 409)"
echo "=========================================================="

curl -s -w "\nHTTP Status: %{http_code}\n" -X POST "$BASE_URL/api/user/register" \
  -H "Content-Type: application/json" \
  -d '{
    "login": "testuser",
    "password": "password123"
  }'
echo ""

# ============================================
# 3. ЛОГИН
# ============================================

echo "🔐 3. Логин существующего пользователя"
echo "======================================="

LOGIN_RESPONSE=$(curl -s -i -X POST "$BASE_URL/api/user/login" \
  -H "Content-Type: application/json" \
  -d '{
    "login": "testuser",
    "password": "password123"
  }')

echo "$LOGIN_RESPONSE"
echo ""

# ============================================
# 4. НЕВЕРНЫЙ ЛОГИН (401)
# ============================================

echo "🔐 4. Неверный пароль (должен вернуть 401)"
echo "=========================================="

curl -s -w "\nHTTP Status: %{http_code}\n" -X POST "$BASE_URL/api/user/login" \
  -H "Content-Type: application/json" \
  -d '{
    "login": "testuser",
    "password": "wrongpassword"
  }'
echo ""

# ============================================
# 5. ЗАГРУЗКА ЗАКАЗА (валидный по Луну)
# ============================================

echo "📦 5. Загрузка заказа (валидный номер по Луну)"
echo "=============================================="

curl -s -w "\nHTTP Status: %{http_code}\n" -X POST "$BASE_URL/api/user/orders" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: text/plain" \
  -d "79927398713"
echo ""

# ============================================
# 6. ПОВТОРНАЯ ЗАГРУЗКА ТОГО ЖЕ ЗАКАЗА (200)
# ============================================

echo "📦 6. Повторная загрузка того же заказа (должен вернуть 200)"
echo "============================================================="

curl -s -w "\nHTTP Status: %{http_code}\n" -X POST "$BASE_URL/api/user/orders" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: text/plain" \
  -d "79927398713"
echo ""

# ============================================
# 7. ЗАГРУЗКА НЕВАЛИДНОГО НОМЕРА (422)
# ============================================

echo "📦 7. Загрузка невалидного номера по Луну (должен вернуть 422)"
echo "==============================================================="

curl -s -w "\nHTTP Status: %{http_code}\n" -X POST "$BASE_URL/api/user/orders" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: text/plain" \
  -d "12345678900"
echo ""

# ============================================
# 8. ЗАГРУЗКА ЕЩЁ НЕСКОЛЬКИХ ЗАКАЗОВ
# ============================================

echo "📦 8. Загрузка ещё нескольких заказов"
echo "======================================"

ORDER_NUMBERS=("12345678903" "49927398716" "1234567812")

for ORDER in "${ORDER_NUMBERS[@]}"; do
    echo "Загружаем заказ: $ORDER"
    curl -s -w "\nHTTP Status: %{http_code}\n" -X POST "$BASE_URL/api/user/orders" \
      -H "Authorization: Bearer $TOKEN" \
      -H "Content-Type: text/plain" \
      -d "$ORDER"
    echo ""
done


# Добавляем задержку в 60 секунд
echo "⏳ Ожидание 60 секунд для обработки заказов..."
sleep 60
# ============================================
# 9. ПОЛУЧЕНИЕ СПИСКА ЗАКАЗОВ
# ============================================

echo "📋 9. Получение списка заказов"
echo "==============================="

sleep 2  # Даём время на обработку

curl -s -X GET "$BASE_URL/api/user/orders" \
  -H "Authorization: Bearer $TOKEN" | jq .
echo ""

# ============================================
# 10. ПОЛУЧЕНИЕ БАЛАНСА
# ============================================

echo "💰 10. Получение баланса"
echo "========================"

curl -s -X GET "$BASE_URL/api/user/balance" \
  -H "Authorization: Bearer $TOKEN" | jq .
echo ""

# ============================================
# 11. СПИСАНИЕ БАЛЛОВ
# ============================================

echo "💸 11. Списание баллов"
echo "======================"

curl -s -w "\nHTTP Status: %{http_code}\n" -X POST "$BASE_URL/api/user/balance/withdraw" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "order": "2377225624",
    "sum": 100
  }'
echo ""

# ============================================
# 12. СПИСАНИЕ БОЛЬШЕ ЧЕМ ЕСТЬ (402)
# ============================================

echo "💸 12. Списание больше чем есть на счету (должен вернуть 402)"
echo "=============================================================="

curl -s -w "\nHTTP Status: %{http_code}\n" -X POST "$BASE_URL/api/user/balance/withdraw" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "order": "49927398716",
    "sum": 999999
  }'
echo ""

# ============================================
# 13. ПОЛУЧЕНИЕ ИСТОРИИ СПИСАНИЙ
# ============================================

echo "📊 13. Получение истории списаний"
echo "=================================="

curl -s -X GET "$BASE_URL/api/user/withdrawals" \
  -H "Authorization: Bearer $TOKEN" | jq .
echo ""

# ============================================
# 14. ЗАПРОС БЕЗ ТОКЕНА (401)
# ============================================

echo "🔒 14. Запрос без токена (должен вернуть 401)"
echo "=============================================="

curl -s -w "\nHTTP Status: %{http_code}\n" -X GET "$BASE_URL/api/user/orders"
echo ""

# ============================================
# 15. ЗАПРОС С НЕВАЛИДНЫМ ТОКЕНОМ (401)
# ============================================

echo "🔒 15. Запрос с невалидным токеном (должен вернуть 401)"
echo "========================================================"

curl -s -w "\nHTTP Status: %{http_code}\n" -X GET "$BASE_URL/api/user/orders" \
  -H "Authorization: Bearer invalid.token.here"
echo ""

# ============================================
# 16. ВТОРОЙ ПОЛЬЗОВАТЕЛЬ ПЫТАЕТСЯ ЗАГРУЗИТЬ ТОТ ЖЕ ЗАКАЗ (409)
# ============================================

echo "👥 16. Второй пользователь пытается загрузить тот же заказ (409)"
echo "================================================================="

# Регистрируем второго пользователя
USER2_RESPONSE=$(curl -s -i -X POST "$BASE_URL/api/user/register" \
  -H "Content-Type: application/json" \
  -d '{
    "login": "testuser2",
    "password": "password123"
  }')

TOKEN2=$(echo "$USER2_RESPONSE" | grep -i "Authorization:" | awk '{print $3}' | tr -d '\r')

echo "Токен второго пользователя получен"
echo ""

# Пытаемся загрузить заказ первого пользователя
curl -s -w "\nHTTP Status: %{http_code}\n" -X POST "$BASE_URL/api/user/orders" \
  -H "Authorization: Bearer $TOKEN2" \
  -H "Content-Type: text/plain" \
  -d "79927398713"
echo ""

# ============================================
# ИТОГИ
# ============================================

echo "✅ Тестирование завершено!"
echo ""
echo "💡 Подождите 10-30 секунд и проверьте статусы заказов снова:"
echo "   curl -H 'Authorization: Bearer $TOKEN' $BASE_URL/api/user/orders | jq ."
echo ""
echo "📊 Проверьте обновлённый баланс:"
echo "   curl -H 'Authorization: Bearer $TOKEN' $BASE_URL/api/user/balance | jq ."
echo ""
echo "🔑 Сохранённый токен для дальнейшего использования:"
echo "   TOKEN=$TOKEN"