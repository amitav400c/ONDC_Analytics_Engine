const http = require('http');

const GATEWAY_URL = process.env.GATEWAY_URL || 'http://localhost:8080';
const CONCURRENCY = 2000;
const DURATION_MS = 10000;

async function sendRequest(id) {
  return new Promise((resolve) => {
    const payload = JSON.stringify({
      context: { action: "on_search", transaction_id: `txn-${id}` },
      message: { data: "test" }
    });

    const req = http.request(`${GATEWAY_URL}/webhooks/ondc`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        'Content-Length': Buffer.byteLength(payload),
      },
      agent: new http.Agent({ keepAlive: true, maxSockets: Infinity })
    }, (res) => {
      resolve(res.statusCode);
    });

    req.on('error', (err) => {
      resolve(err.message); // Count as error
    });

    req.write(payload);
    req.end();
  });
}

async function run() {
  console.log(`Starting stress test against ${GATEWAY_URL}`);
  console.log(`Concurrency: ${CONCURRENCY}, Duration: ${DURATION_MS}ms`);

  const start = Date.now();
  let completed = 0;
  let successes = 0;
  let failures = 0;

  const runLoop = async (workerId) => {
    while (Date.now() - start < DURATION_MS) {
      const status = await sendRequest(`${workerId}-${completed}`);
      completed++;
      if (status === 202) successes++;
      else failures++;
    }
  };

  const workers = [];
  for (let i = 0; i < CONCURRENCY; i++) {
    workers.push(runLoop(i));
  }

  await Promise.all(workers);

  const durationSec = (Date.now() - start) / 1000;
  const rps = (completed / durationSec).toFixed(2);

  console.log(`\n=== Stress Test Results ===`);
  console.log(`Total Requests: ${completed}`);
  console.log(`Successful (HTTP 202): ${successes}`);
  console.log(`Failed/Rejected: ${failures}`);
  console.log(`Throughput: ${rps} RPS`);
}

run().catch(console.error);
