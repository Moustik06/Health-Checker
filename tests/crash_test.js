import http from 'k6/http';
import { check } from 'k6';

export const options = {
    scenarios: {
        constant_request_rate: {
            executor: 'constant-arrival-rate',
            rate: 500,
            timeUnit: '1s',
            duration: '1m',
            preAllocatedVUs: 100,
            maxVUs: 400,
        },
    },
};

const allUrls = [
    "https://google.com", "https://github.com", "https://gitlab.com",
    "https://microsoft.com", "https://amazon.com", "https://apple.com",
    "https://golang.org", "https://docker.com", "https://kubernetes.io",
    "https://prometheus.io", "https://grafana.com", "https://developer.mozilla.org",
    "https://stackoverflow.com", "https://reddit.com", "https://unexistingsite.error",
    "https://www.wikipedia.org/", "https://www.youtube.com/", "https://www.facebook.com/",
    "https://www.twitter.com/", "https://www.instagram.com/"
];


function getRandomSubarray(arr, size) {
    const shuffled = arr.sort(() => 0.5 - Math.random());
    return shuffled.slice(0, size);
}

const params = {
    headers: { 'Content-Type': 'application/json' },
};

export default function () {
    const numUrls = Math.floor(Math.random() * 6) + 5;
    const urlsToCheck = getRandomSubarray(allUrls, numUrls);

    const payload = JSON.stringify({ urls: urlsToCheck });

    const res = http.post('http://localhost:8080/check', payload, params);
    check(res, { 'status was 200': (r) => r.status == 200 });
}