const API_URL = 'http://localhost:5000/api';

function getUserId() {
    const userId = document.getElementById('userId').value.trim();
    if (!userId) {
        showResult('balanceResult', 'Пожалуйста, введите User ID', 'error');
        return null;
    }
    return userId;
}

function showResult(elementId, message, type = 'success') {
    const element = document.getElementById(elementId);
    element.className = `result ${type}`;
    element.innerHTML = typeof message === 'object'
        ? `<pre>${JSON.stringify(message, null, 2)}</pre>`
        : message;
}

async function apiCall(url, options = {}) {
    try {
        const response = await fetch(url, options);
        const data = await response.json();

        if (!response.ok) {
            throw new Error(data.error || 'Ошибка запроса');
        }

        return { success: true, data };
    } catch (error) {
        return { success: false, error: error.message };
    }
}

async function createAccount() {
    const userId = getUserId();
    if (!userId) return;

    const result = await apiCall(`${API_URL}/accounts`, {
        method: 'POST',
        headers: {
            'Content-Type': 'application/json',
            'X-User-Id': userId
        }
    });

    if (result.success) {
        showResult('balanceResult', `✅ Аккаунт создан!\n${JSON.stringify(result.data, null, 2)}`, 'success');
    } else {
        showResult('balanceResult', `❌ Ошибка: ${result.error}`, 'error');
    }
}

async function getBalance() {
    const userId = getUserId();
    if (!userId) return;

    const result = await apiCall(`${API_URL}/accounts/balance`, {
        headers: {
            'X-User-Id': userId
        }
    });

    if (result.success) {
        showResult('balanceResult', `💰 Баланс: ${result.data.balance} руб.\n${JSON.stringify(result.data, null, 2)}`, 'success');
    } else {
        showResult('balanceResult', `❌ Ошибка: ${result.error}`, 'error');
    }
}

async function deposit() {
    const userId = getUserId();
    if (!userId) return;

    const amount = parseFloat(document.getElementById('depositAmount').value);
    if (!amount || amount <= 0) {
        showResult('balanceResult', 'Введите корректную сумму', 'error');
        return;
    }

    const result = await apiCall(`${API_URL}/accounts/deposit`, {
        method: 'POST',
        headers: {
            'Content-Type': 'application/json',
            'X-User-Id': userId
        },
        body: JSON.stringify({ amount })
    });

    if (result.success) {
        showResult('balanceResult', `✅ Пополнение успешно!\nНовый баланс: ${result.data.balance} руб.`, 'success');
        document.getElementById('depositAmount').value = '';
    } else {
        showResult('balanceResult', `❌ Ошибка: ${result.error}`, 'error');
    }
}

async function createOrder() {
    const userId = getUserId();
    if (!userId) return;

    const amount = parseFloat(document.getElementById('orderAmount').value);
    const description = document.getElementById('orderDescription').value.trim();

    if (!amount || amount <= 0) {
        showResult('ordersResult', 'Введите корректную сумму заказа', 'error');
        return;
    }

    if (!description) {
        showResult('ordersResult', 'Введите описание заказа', 'error');
        return;
    }

    const result = await apiCall(`${API_URL}/orders`, {
        method: 'POST',
        headers: {
            'Content-Type': 'application/json',
            'X-User-Id': userId
        },
        body: JSON.stringify({ amount, description })
    });

    if (result.success) {
        showResult('ordersResult', `✅ Заказ создан!\n${JSON.stringify(result.data, null, 2)}`, 'success');
        document.getElementById('orderAmount').value = '';
        document.getElementById('orderDescription').value = '';

        setTimeout(() => getOrders(), 2000);
    } else {
        showResult('ordersResult', `❌ Ошибка: ${result.error}`, 'error');
    }
}

async function getOrders() {
    const userId = getUserId();
    if (!userId) return;

    const result = await apiCall(`${API_URL}/orders`, {
        headers: {
            'X-User-Id': userId
        }
    });

    if (result.success) {
        if (result.data.length === 0) {
            showResult('ordersResult', 'У вас пока нет заказов', 'success');
        } else {
            const ordersHtml = result.data.map(order => `
                <div class="order-item ${order.status.toLowerCase()}">
                    <h3>📦 Заказ #${order.id.substring(0, 8)}</h3>
                    <p><strong>Описание:</strong> ${order.description}</p>
                    <p><strong>Сумма:</strong> ${order.amount} руб.</p>
                    <p><strong>Статус:</strong> ${order.status}</p>
                    <p><strong>Создан:</strong> ${new Date(order.created_at).toLocaleString('ru-RU')}</p>
                </div>
            `).join('');

            document.getElementById('ordersResult').innerHTML = ordersHtml;
            document.getElementById('ordersResult').className = 'result';
        }
    } else {
        showResult('ordersResult', `❌ Ошибка: ${result.error}`, 'error');
    }
}

async function checkHealth() {
    const result = await apiCall(`${API_URL.replace('/api', '')}/health`);

    if (result.success) {
        showResult('healthResult', `✅ Система работает нормально\n${JSON.stringify(result.data, null, 2)}`, 'success');
    } else {
        showResult('healthResult', `❌ Система недоступна: ${result.error}`, 'error');
    }
}

// Set default user ID on load
window.addEventListener('DOMContentLoaded', () => {
    document.getElementById('userId').value = '550e8400-e29b-41d4-a716-446655440000';
});
