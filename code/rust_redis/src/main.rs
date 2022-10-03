use rdkafka::consumer::stream_consumer::StreamConsumer;
use rdkafka::consumer::{CommitMode, Consumer};
use rdkafka::{ClientConfig, Message};
use rdkafka::config::RDKafkaLogLevel;
use simple_redis;
use std::boxed::Box;
use log::{error, info, warn};
use serde::{Deserialize, Serialize};
use simple_redis::client::Client;

#[derive(Serialize, Deserialize)]
pub struct Answer {
    #[serde(rename = "payload")]
    payload: Payload,
}

#[derive(Serialize, Deserialize)]
pub struct Payload {
    #[serde(rename = "before")]
    before: Option<serde_json::Value>,

    #[serde(rename = "after")]
    after: After,

    #[serde(rename = "source")]
    source: Source,

    #[serde(rename = "op")]
    op: String,

    #[serde(rename = "ts_ms")]
    ts_ms: i64,

    #[serde(rename = "transaction")]
    transaction: Option<serde_json::Value>,
}

#[derive(Serialize, Deserialize)]
pub struct After {
    #[serde(rename = "calldate")]
    calldate: String,

    #[serde(rename = "clid")]
    clid: String,

    #[serde(rename = "src")]
    src: String,

    #[serde(rename = "dst")]
    dst: String,

    #[serde(rename = "dcontext")]
    dcontext: String,

    #[serde(rename = "channel")]
    channel: String,

    #[serde(rename = "dstchannel")]
    dstchannel: String,

    #[serde(rename = "lastapp")]
    lastapp: String,

    #[serde(rename = "lastdata")]
    lastdata: String,

    #[serde(rename = "duration")]
    duration: i64,

    #[serde(rename = "billsec")]
    billsec: i64,

    #[serde(rename = "disposition")]
    disposition: String,

    #[serde(rename = "amaflags")]
    amaflags: i64,

    #[serde(rename = "accountcode")]
    accountcode: String,

    #[serde(rename = "uniqueid")]
    uniqueid: String,

    #[serde(rename = "userfield")]
    userfield: String,

    #[serde(rename = "did")]
    did: String,

    #[serde(rename = "recordingfile")]
    recordingfile: String,

    #[serde(rename = "cnum")]
    cnum: String,

    #[serde(rename = "cnam")]
    cnam: String,

    #[serde(rename = "outbound_cnum")]
    outbound_cnum: String,

    #[serde(rename = "outbound_cnam")]
    outbound_cnam: String,

    #[serde(rename = "dst_cnam")]
    dst_cnam: String,
}

#[derive(Serialize, Deserialize)]
pub struct Source {
    #[serde(rename = "version")]
    version: String,

    #[serde(rename = "connector")]
    connector: String,

    #[serde(rename = "name")]
    name: String,

    #[serde(rename = "ts_ms")]
    ts_ms: i64,

    #[serde(rename = "snapshot")]
    snapshot: String,

    #[serde(rename = "db")]
    db: String,

    #[serde(rename = "sequence")]
    sequence: Option<serde_json::Value>,

    #[serde(rename = "table")]
    table: String,

    #[serde(rename = "server_id")]
    server_id: i64,

    #[serde(rename = "gtid")]
    gtid: String,

    #[serde(rename = "file")]
    file: String,

    #[serde(rename = "pos")]
    pos: i64,

    #[serde(rename = "row")]
    row: i64,

    #[serde(rename = "thread")]
    thread: Option<serde_json::Value>,

    #[serde(rename = "query")]
    query: Option<serde_json::Value>,
}

async fn worker(consumer: StreamConsumer, mut connect_redis: Client) {
    loop {
        match consumer.recv().await {
            Err(e) => warn!("Kafka error: {}", e),
            Ok(m) => {
                let payload = match m.payload_view::<str>() {
                    None => "",
                    Some(Ok(s)) => {
                        info!("{:?}", s);
                        s
                    },
                    Some(Err(e)) => {
                        warn!("Error while deserializing message payload: {:?}", e);
                        ""
                    }
                };
                let cdc_struct: Answer = serde_json::from_str(payload).unwrap();
                let did = cdc_struct.payload.after.did.as_str();
                let billsec = cdc_struct.payload.after.billsec;
                let context = cdc_struct.payload.after.dcontext.as_str();
                connect_redis.hset(context, did, billsec).unwrap();
                consumer.commit_message(&m, CommitMode::Async).unwrap();
            }
        };
    }
}

#[tokio::main]
async fn main() {
    match env_logger::try_init(){
        Err(_err) => {
            error!("Logger init error");
        }
        _ => {
            info!("Logger init");
        }
    };
    let mut connect_redis = simple_redis::create("redis://host-of-redis:6379/").unwrap();

    let consumer_redpanda: StreamConsumer = ClientConfig::new()
        .set("bootstrap.servers", "host-of-redpanda:9092")
        .set("group.id", "mysql-cdr")
        .set("enable.partition.eof", "false")
        .set_log_level(RDKafkaLogLevel::Debug)
        .create()
        .expect("Consumer creation failed");

    consumer_redpanda.subscribe(&vec!["mysql-cdr.asteriskcdrdb.bla-bla".as_ref()]).expect("Can't subscribe to specified topics");
    worker(consumer_redpanda, connect_redis).await;
}
