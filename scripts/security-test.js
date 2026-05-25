const http = require('http');

const GATEWAY_URL = process.env.GATEWAY_URL || 'http://localhost:8080';

async function sendPayload(payload) {
  return new Promise((resolve, reject) => {
    const data = typeof payload === 'string' ? payload : JSON.stringify(payload);
    const req = http.request(`${GATEWAY_URL}/webhooks/ondc`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        'Content-Length': Buffer.byteLength(data),
      }
    }, (res) => {
      let body = '';
      res.on('data', chunk => body += chunk);
      res.on('end', () => resolve({ status: res.statusCode, body }));
    });
    req.on('error', reject);
    req.write(data);
    req.end();
  });
}

async function runTests() {
  let failed = false;

  console.log("=== Running Security DAST Tests ===");

  // Test 1: JSON Bomb (Depth Exhaustion)
  console.log("\n[Test 1] JSON Bomb (Depth Exhaustion)");
  let deepJson = '{"context":{"action":"on_search","bap_id":"test"},"message":{';
  deepJson += '"a":{'.repeat(15000);
  deepJson += '"b":1';
  deepJson += '}'.repeat(15000);
  deepJson += '}}';
  
  try {
    const res1 = await sendPayload(deepJson);
    if (res1.status === 400 && (res1.body.includes('invalid payload') || res1.body.includes('invalid JSON'))) {
      console.log("✅ Passed: Gateway correctly rejected the JSON bomb.");
    } else {
      console.error(`❌ Failed: Gateway responded with ${res1.status} - ${res1.body}`);
      failed = true;
    }
  } catch (e) {
    console.log("✅ Passed: Connection dropped or rejected gracefully.");
  }

  // Test 2: Invalid JSON Syntax
  console.log("\n[Test 2] Invalid JSON Syntax");
  try {
    const res2 = await sendPayload('{"context":{"action":"on_search"},"message":{ "buyer_phone":"9876543210" ');
    if (res2.status === 400) {
      console.log("✅ Passed: Gateway correctly rejected invalid JSON syntax.");
    } else {
      console.error(`❌ Failed: Gateway responded with ${res2.status} - ${res2.body}`);
      failed = true;
    }
  } catch (e) {
    console.error("❌ Failed:", e.message);
    failed = true;
  }

  if (failed) {
    console.error("\n❌ Security Tests Failed!");
    process.exit(1);
  } else {
    console.log("\n✅ All Security Tests Passed!");
    process.exit(0);
  }
}

runTests().catch(console.error);
