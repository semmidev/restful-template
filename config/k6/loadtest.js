/*
Usage:
  BASE_URL=http://localhost:8080 k6 run config/k6/loadtest.js

Environment variables:
  BASE_URL  — API base URL (default: http://localhost:8080)
*/

import http from 'k6/http';
import { check, sleep, group } from 'k6';
import { uuidv4 } from 'https://jslib.k6.io/k6-utils/1.4.0/index.js';

// ─── Load Test Configuration ──────────────────────────────────────────────────
export const options = {
    stages: [
        { duration: '2m', target: 25  }, // ramp up to 25 VUs
        { duration: '1m', target: 0    }, // ramp down to 0 VUs
    ],
    thresholds: {
        http_req_duration: ['p(95)<500'], // 95% of requests must complete under 500ms
        http_req_failed:   ['rate<0.01'], // error rate must stay under 1%
    },
};

// ─── Helpers ──────────────────────────────────────────────────────────────────
const BASE_URL = __ENV.BASE_URL || 'http://localhost:8080';

const JSON_HEADERS = { 'Content-Type': 'application/json' };

/** Generate a random alphanumeric string of the given length. */
function randomString(length) {
    const chars = 'abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789';
    let result = '';
    for (let i = 0; i < length; i++) {
        result += chars.charAt(Math.floor(Math.random() * chars.length));
    }
    return result;
}

// ─── Default scenario ─────────────────────────────────────────────────────────
export default function () {
    // Each VU iteration uses a unique email so registrations never conflict.
    const email    = `loadtest-${uuidv4()}@example.com`;
    const password = `P@ss-${randomString(12)}`;

    // ── 1. Health Check ───────────────────────────────────────────────────────
    group('Health Check', function () {
        const res = http.get(`${BASE_URL}/api/v1/health`, {
            tags: { name: 'HealthCheck' },
        });
        check(res, {
            'health: status 200': (r) => r.status === 200,
        });
    });

    sleep(0.5);

    // ── 2. Register ───────────────────────────────────────────────────────────
    let hasAuth = false;

    group('Register', function () {
        const res = http.post(
            `${BASE_URL}/api/v1/auth/register`,
            JSON.stringify({ email, password }),
            { headers: JSON_HEADERS, tags: { name: 'Register' } },
        );

        const ok = check(res, {
            'register: status 200': (r) => r.status === 200 || r.status === 201,
            'register: has access_token cookie': (r) => !!r.cookies.access_token,
        });

        if (ok) {
            hasAuth = true;
        }
    });

    sleep(0.5);

    // ── 3. Login (re-authenticate the just-registered user) ───────────────────
    group('Login', function () {
        const res = http.post(
            `${BASE_URL}/api/v1/auth/login`,
            JSON.stringify({ email, password }),
            { headers: JSON_HEADERS, tags: { name: 'Login' } },
        );

        const ok = check(res, {
            'login: status 200': (r) => r.status === 200,
            'login: has access_token cookie': (r) => !!r.cookies.access_token,
        });

        if (ok) {
            hasAuth = true;
        }
    });

    sleep(0.5);
    
    // Stop if we don't have authentication cookies
    if (!hasAuth) return;

    // ── 4. Create Todo ────────────────────────────────────────────────────────
    let todoId = null;
    group('Create Todo', function () {
        const res = http.post(
            `${BASE_URL}/api/v1/todos`,
            JSON.stringify({ title: 'Load Test Todo', description: 'Created during k6 load test', status: 'pending' }),
            { headers: JSON_HEADERS, tags: { name: 'CreateTodo' } },
        );

        const ok = check(res, {
            'create todo: status 201': (r) => r.status === 201,
            'create todo: has id': (r) => {
                try { return !!JSON.parse(r.body).id; } catch (e) { return false; }
            },
        });

        if (ok) {
            todoId = JSON.parse(res.body).id;
        }
    });

    sleep(0.5);
    if (!todoId) return;

    // ── 5. Get Todo ───────────────────────────────────────────────────────────
    group('Get Todo', function () {
        const res = http.get(
            `${BASE_URL}/api/v1/todos/${todoId}`,
            { headers: JSON_HEADERS, tags: { name: 'GetTodo' } },
        );

        check(res, {
            'get todo: status 200': (r) => r.status === 200,
            'get todo: id matches': (r) => {
                try { return JSON.parse(r.body).id === todoId; } catch (e) { return false; }
            },
        });
    });

    sleep(0.5);

    // ── 6. List Todos ─────────────────────────────────────────────────────────
    group('List Todos', function () {
        const res = http.get(
            `${BASE_URL}/api/v1/todos?page=1&per_page=10`,
            { headers: JSON_HEADERS, tags: { name: 'ListTodos' } },
        );

        check(res, {
            'list todos: status 200': (r) => r.status === 200,
            'list todos: is array': (r) => {
                try { return Array.isArray(JSON.parse(r.body)); } catch (e) { return false; }
            },
        });
    });

    sleep(0.5);

    // ── 7. Delete Todo ────────────────────────────────────────────────────────
    group('Delete Todo', function () {
        const res = http.del(
            `${BASE_URL}/api/v1/todos/${todoId}`,
            null,
            { headers: JSON_HEADERS, tags: { name: 'DeleteTodo' } },
        );

        check(res, {
            'delete todo: status 204': (r) => r.status === 204,
        });
    });

    sleep(1);
}
