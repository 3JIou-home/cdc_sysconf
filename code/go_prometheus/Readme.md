# Пример реализации коннектора Cdc -> Redpanda -> Prometheus на Go

## Описание
Данный пример показывает как можно реализовать коннектор между Cdc и Prometheus. 
В качестве источника данных используется Redpanda, а в качестве конечной точки симулируется Prometheus exporter. 
В качестве примера используется топик `mysql-cdr.asteriskcdrdb.bal-bal` с json представленным [как пример](https://github.com/3JIou-home/cdc_sysconf/blob/3aa604b63af6745f32967f2d37ca8d1a20536942/code/cdc_json_example.json) парсим код оператора и считаем количество звонков по каждому оператору - транку.

## Запуск

```bash
go run main.go
```