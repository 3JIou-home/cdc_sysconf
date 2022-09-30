# BINARY BUILD STAGE
FROM ubuntu:20.04 as builder
WORKDIR /
RUN apt update &&\
    apt install -y wget &&\
    wget https://github.com/scylladb/scylla-cdc-source-connector/releases/download/scylla-cdc-source-connector-1.0.1/scylla-cdc-source-connector-1.0.1-jar-with-dependencies.jar &&\
    wget https://repo1.maven.org/maven2/io/prometheus/jmx/jmx_prometheus_javaagent/0.16.1/jmx_prometheus_javaagent-0.16.1.jar

FROM debezium/connect
WORKDIR /kafka/connect

ENV TZ=Asia/Almaty
USER root

COPY --from=builder  /scylla-cdc-source-connector-1.0.1-jar-with-dependencies.jar ./debezium-connector-scylladb/scylla-cdc-connector.jar
COPY --from=builder  /jmx_prometheus_javaagent-0.16.1.jar ../etc/jmx_prometheus_javaagent.jar
COPY ./jmx_config.yml ../etc/config.yml
RUN chown -R kafka: ./debezium-connector-scylladb &&\
    chown -R kafka: ../etc

# Metadata
LABEL org.kolesa-team.image.maintainer="Vlassov Dmitriy <vlassov@kolesa.kz>" \
      org.kolesa-team.image.description="Debezium with scylladb-cdc connector."
