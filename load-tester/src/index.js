#!/usr/bin/env node

// ONDC Load Tester — Generates realistic Beckn protocol payloads with synthetic PII.
// Usage: GATEWAY_URL=http://localhost:8080 RPS=50 DURATION_SECONDS=30 node src/index.js

const GATEWAY_URL = process.env.GATEWAY_URL || 'http://localhost:8080';
const RPS = parseInt(process.env.RPS || '50', 10);
const DURATION = parseInt(process.env.DURATION_SECONDS || '30', 10);

const ACTIONS = ['on_search', 'on_select', 'on_init', 'on_confirm', 'on_cancel'];
const ACTION_WEIGHTS = [40, 25, 15, 12, 8]; // Realistic funnel distribution

const CITIES = [
  { code: 'BLR', gps: '12.9716,77.5946' },
  { code: 'DEL', gps: '28.7041,77.1025' },
  { code: 'MUM', gps: '19.0760,72.8777' },
  { code: 'HYD', gps: '17.3850,78.4867' },
  { code: 'CHE', gps: '13.0827,80.2707' },
  { code: 'PUN', gps: '18.5204,73.8567' },
  { code: 'KOL', gps: '22.5726,88.3639' },
];

const DOMAINS = ['ONDC:RET10', 'ONDC:RET11', 'ONDC:RET12', 'ONDC:RET13'];

function randomPhone() {
  return '9' + Math.floor(100000000 + Math.random() * 900000000).toString();
}

function randomId() {
  return Array.from({ length: 16 }, () =>
    Math.floor(Math.random() * 16).toString(16)
  ).join('');
}

function weightedRandom(items, weights) {
  const total = weights.reduce((a, b) => a + b, 0);
  let r = Math.random() * total;
  for (let i = 0; i < items.length; i++) {
    r -= weights[i];
    if (r <= 0) return items[i];
  }
  return items[items.length - 1];
}

function generatePayload() {
  const action = weightedRandom(ACTIONS, ACTION_WEIGHTS);
  const city = CITIES[Math.floor(Math.random() * CITIES.length)];
  const txnId = randomId();

  return {
    context: {
      action,
      domain: DOMAINS[Math.floor(Math.random() * DOMAINS.length)],
      city: city.code,
      transaction_id: txnId,
      message_id: randomId(),
      timestamp: new Date().toISOString(),
      bap_id: 'buyer-app-' + Math.floor(Math.random() * 5),
      bpp_id: 'seller-app-' + Math.floor(Math.random() * 10),
    },
    message: {
      order: {
        id: txnId,
        state: action === 'on_cancel' ? 'Cancelled' : 'Active',
        provider: {
          id: 'seller-' + Math.floor(Math.random() * 100),
          locations: [{ gps: city.gps }],
        },
        items: [
          {
            id: 'item-' + Math.floor(Math.random() * 1000),
            quantity: { count: Math.floor(Math.random() * 5) + 1 },
            price: { value: (Math.random() * 2000 + 50).toFixed(2), currency: 'INR' },
          },
        ],
        billing: {
          name: 'Test Buyer',
          phone: randomPhone(),
          address: { city: city.code, gps: city.gps },
        },
        fulfillment: {
          type: 'Delivery',
          end: { location: { gps: city.gps } },
        },
        payment: {
          type: 'ON-ORDER',
          status: 'PAID',
          params: { amount: (Math.random() * 2000 + 50).toFixed(2), currency: 'INR' },
        },
      },
    },
  };
}

async function sendRequest() {
  const payload = generatePayload();
  try {
    const res = await fetch(`${GATEWAY_URL}/webhooks/ondc`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(payload),
    });
    return { status: res.status, action: payload.context.action };
  } catch (err) {
    return { status: 0, error: err.message };
  }
}

async function main() {
  console.log(`\n🚀 ONDC Load Tester`);
  console.log(`   Target:   ${GATEWAY_URL}`);
  console.log(`   RPS:      ${RPS}`);
  console.log(`   Duration: ${DURATION}s`);
  console.log(`   Total:    ~${RPS * DURATION} requests\n`);

  const intervalMs = 1000 / RPS;
  let sent = 0, success = 0, failed = 0;
  const startTime = Date.now();
  const endTime = startTime + DURATION * 1000;

  const stats = {};
  ACTIONS.forEach(a => (stats[a] = 0));

  const interval = setInterval(async () => {
    if (Date.now() >= endTime) {
      clearInterval(interval);
      const elapsed = ((Date.now() - startTime) / 1000).toFixed(1);
      console.log(`\n✅ Complete — ${sent} sent, ${success} ok, ${failed} failed in ${elapsed}s`);
      console.log(`   Effective RPS: ${(sent / parseFloat(elapsed)).toFixed(1)}`);
      console.log(`   Distribution:`, stats);
      process.exit(0);
    }

    const result = await sendRequest();
    sent++;
    if (result.status === 202) {
      success++;
      if (result.action) stats[result.action]++;
    } else {
      failed++;
    }

    if (sent % 100 === 0) {
      process.stdout.write(`\r   Sent: ${sent} | OK: ${success} | Fail: ${failed}`);
    }
  }, intervalMs);
}

main().catch(console.error);
