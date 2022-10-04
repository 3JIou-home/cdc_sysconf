# Пример реализации коннектора Cdc -> Redpanda -> Redis на Rust

## Описание
Данный пример показывает как можно реализовать коннектор между Cdc и Redis. 
В качестве источника данных используется Redpanda, а в качестве конечной точки Redis. 
В качестве примера используется топик `mysql-cdr.asteriskcdrdb.bal-bal` с json представленным [как пример](https://github.com/3JIou-home/cdc_sysconf/blob/3aa604b63af6745f32967f2d37ca8d1a20536942/code/cdc_json_example.json) и несколькими значениями `context`, `billsec` и `did`.

## Запуск

```bash
cargo run
```