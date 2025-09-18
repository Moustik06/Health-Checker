import http from 'k6/http';
import { check, sleep } from 'k6';


export const options = {
    vus: 10,
    duration: '30s',
};


const payload = JSON.stringify({
    urls: [
        'https://test.k6.io',
        'https://httpbin.org/delay/1',
        'https://invalid.url.that.does.not.exist',
    ],
});

const params = {
    headers: {
        'Content-Type': 'application/json',
    },
};

export default function () {
    const res = http.post('http://localhost:8080/check', payload, params);

    check(res, { 'status was 200': (r) => r.status == 200 });
    sleep(1);
}