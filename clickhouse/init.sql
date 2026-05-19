-- ONDC Analytics Gateway — ClickHouse Schema
-- Auto-ingests from Redpanda topic via Kafka Engine

CREATE DATABASE IF NOT EXISTS ondc;

-- Kafka Engine table: pulls raw events from Redpanda
-- Uses String for timestamp to handle ISO 8601 format from Go
CREATE TABLE IF NOT EXISTS ondc.ondc_events_queue (
    event_id     String,
    event_type   String,
    action       String,
    city         String,
    timestamp    String,
    order_id     String,
    seller_id    String,
    buyer_hash   String,
    gps_lat      Float64,
    gps_lng      Float64,
    amount       Float64,
    status       String,
    domain       String,
    raw_payload  String
) ENGINE = Kafka
SETTINGS
    kafka_broker_list = 'redpanda:9092',
    kafka_topic_list = 'ondc_events_sanitized',
    kafka_group_name = 'clickhouse_consumer',
    kafka_format = 'JSONEachRow',
    kafka_num_consumers = 1;

-- MergeTree table: the actual queryable store
CREATE TABLE IF NOT EXISTS ondc.ondc_events (
    event_id     String,
    event_type   String,
    action       String,
    city         String,
    timestamp    DateTime64(3),
    order_id     String,
    seller_id    String,
    buyer_hash   String,
    gps_lat      Float64,
    gps_lng      Float64,
    amount       Float64,
    status       String,
    domain       String,
    raw_payload  String,
    inserted_at  DateTime DEFAULT now()
) ENGINE = MergeTree()
ORDER BY (event_type, city, timestamp)
PARTITION BY toYYYYMM(timestamp);

-- Materialized View: routes Kafka queue → MergeTree
-- Parses ISO 8601 timestamp string into DateTime64
CREATE MATERIALIZED VIEW IF NOT EXISTS ondc.ondc_events_mv
TO ondc.ondc_events AS
SELECT
    event_id,
    event_type,
    action,
    city,
    parseDateTimeBestEffort(timestamp) AS timestamp,
    order_id,
    seller_id,
    buyer_hash,
    gps_lat,
    gps_lng,
    amount,
    status,
    domain,
    raw_payload
FROM ondc.ondc_events_queue;

-- Aggregation view: funnel metrics
CREATE VIEW IF NOT EXISTS ondc.funnel_metrics AS
SELECT
    event_type,
    count() AS event_count,
    uniqExact(order_id) AS unique_orders,
    toStartOfDay(timestamp) AS day
FROM ondc.ondc_events
GROUP BY event_type, day
ORDER BY day DESC, event_type;

-- Aggregation view: cancellations by city
CREATE VIEW IF NOT EXISTS ondc.cancellation_metrics AS
SELECT
    city,
    count() AS cancel_count,
    toStartOfDay(timestamp) AS day
FROM ondc.ondc_events
WHERE event_type = 'on_cancel'
GROUP BY city, day
ORDER BY cancel_count DESC;
