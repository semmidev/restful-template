/*
Usage:
DOC_API_BASE_URL=<URL> k6 run loadtest.js
*/

import http from 'k6/http';
import { check, sleep, group } from 'k6';

// Load Test Configuration
export const options = {
    stages: [
        { duration: '2m', target: 1000 }, // ramp up to 1000 VUs
        { duration: '5m', target: 1500 }, // hold 1000 VUs
        { duration: '2m', target: 0 },    // ramp down to 0 VUs
    ],
    thresholds: {
        http_req_duration: ['p(95)<500'], // 95% of requests must be under 500ms
        http_req_failed: ['rate<0.01'],   // error rate must be under 1%
    },
};

// Configure BASE_URL from environment variable
const BASE_URL = __ENV.DOC_API_BASE_URL || 'http://localhost:8080';

// In-memory file content (simple byte array)
const fileContent = new Uint8Array([0x68, 0x65, 0x6c, 0x6c, 0x6f]).buffer;

export default function () {
    let uploadedDocId = null;

    // Health Check
    group('Health Check', function () {
        const res = http.get(`${BASE_URL}/api/v1/health`, {
            tags: { name: 'HealthCheck' },
        });

        check(res, {
            'health status is 200': (r) => r.status === 200,
        });
    });

    // End of iteration
    sleep(1);
}
