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
        { duration: '2m', target: 100  }, // ramp up to 500 VUs
        // { duration: '2m', target: 1000 }, // ramp up to 1000 VUs
        // { duration: '1m', target: 2000 }, // ramp up to 2000 VUs
        // { duration: '1m', target: 2000 }, // hold at 2000 VUs
        // { duration: '1m', target: 1000 }, // ramp down to 1000 VUs
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
    // group('Health Check', function () {
    //     const res = http.get(`${BASE_URL}/api/v1/health`, {
    //         tags: { name: 'HealthCheck' },
    //     });
    //     check(res, {
    //         'health: status 200': (r) => r.status === 200,
    //     });
    // });

    // sleep(0.5);

    // ── 2. Register ───────────────────────────────────────────────────────────
    let accessToken = null;

    group('Register', function () {
        const res = http.post(
            `${BASE_URL}/api/v1/auth/register`,
            JSON.stringify({ email, password }),
            { headers: JSON_HEADERS, tags: { name: 'Register' } },
        );

        const ok = check(res, {
            'register: status 201': (r) => r.status === 201,
            'register: has access_token': (r) => {
                try { return !!JSON.parse(r.body).access_token; } catch (e) { return false; }
            },
        });

        if (ok) {
            accessToken = JSON.parse(res.body).access_token;
        }
    });

    sleep(0.5);

    // ── 3. Login (re-authenticate the just-registered user) ───────────────────
    // group('Login', function () {
    //     const res = http.post(
    //         `${BASE_URL}/api/v1/auth/login`,
    //         JSON.stringify({ email, password }),
    //         { headers: JSON_HEADERS, tags: { name: 'Login' } },
    //     );

    //     check(res, {
    //         'login: status 200': (r) => r.status === 200,
    //         'login: has access_token': (r) => {
    //             try { return !!JSON.parse(r.body).access_token; } catch (e) { return false; }
    //         },
    //     });

    //     if (!accessToken) {
    //         try { accessToken = JSON.parse(res.body).access_token; } catch (e) { /* ignore */ }
    //     }
    // });

    sleep(1);
}
