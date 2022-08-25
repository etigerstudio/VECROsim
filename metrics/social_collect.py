from prometheus_api_client import PrometheusConnect
from prometheus_api_client.utils import parse_datetime
import string
import csv
import pandas as pd

prometheus_host_url = "http://127.0.0.1:9091/"
start_time = parse_datetime("2022-05-22 21:00:00")
end_time = parse_datetime("2022-05-22 22:00:00")
step = "1s"
filepath = "social-delay/jitter_high"

prom = PrometheusConnect(url=prometheus_host_url, disable_ssl=True)

target = [
    "social-follow-user",
    "social-recommender",
    "social-unique-id",
    "social-url-shorten",
    "social-video",
    "social-image",
    "social-text",
    "social-user-tag",
    "social-favorite",
    "social-search",
    "social-ads",
    "social-read-post",
    "social-login",
    "social-compose-post",
    "social-blocked-users",
    "social-read-timeline",
    "social-user-info",
    "social-posts-storage",
    "social-write-timeline",
    "social-write-graph",
    "social-read-timeline-db",
    "social-user-info-db",
    "social-posts-storage-db",
    "social-write-timeline-db",
    "social-write-graph-db"
]

# now 10s
metrics = {
    "latency_avg": 'rate(ben_base_social_latency_counter{service="$SVC_NAME$"}[10s]) / rate(ben_base_social_request_count{service="$SVC_NAME$"}[10s])',
    "latency_p95": 'histogram_quantile(0.95, rate(ben_base_social_latency_histogram_bucket{service="$SVC_NAME$"}[1m]))',
    "throughput": 'rate(ben_base_social_throughput{service="$SVC_NAME$"}[10s])',
    "cpu_usage": 'rate(container_cpu_usage_seconds_total{pod=~"$SVC_NAME$.*", container="$CONTAINER_NAME$"}[1m])',
    "memory_usage": 'container_memory_working_set_bytes{pod=~"$SVC_NAME$.*", container="$CONTAINER_NAME$"}',
    "network_receive_bytes": 'rate(container_network_receive_bytes_total{pod=~"$SVC_NAME$-[a-z0-9]*-[a-z0-9]*", interface="eth0"}[1m])',
    "network_transmit_bytes": 'rate(container_network_transmit_bytes_total{pod=~"$SVC_NAME$-[a-z0-9]*-[a-z0-9]*",interface="eth0"}[1m])',
}


def target_to_metrics_name(target, m):
    service_name = target
    container_name = target[7:]
    if target.endswith("-db"):
        if m in ["latency_avg", "latency_p95", "throughput"]:
            container_name = f"{container_name}-agent"
        else:
            container_name = f"{container_name}-mongodb"

    query = metrics[m].replace("$SVC_NAME$", target).replace("$CONTAINER_NAME$", container_name)
    print(f"{target}, {m}: {query}")
    return query


for m in metrics.keys():
    rows = {}
    for t in target:
        # query metrics
        query_result_list = prom.custom_query_range(
            target_to_metrics_name(t, m),  # this is the metric name and label config
            start_time=start_time,
            end_time=end_time,
            step=step
        )
        assert len(query_result_list) == 1
        # get metrics values
        metrics_list = query_result_list[0]['values']

        # extract metrics
        row = [m[1] for m in metrics_list]
        series = pd.Series(row)
        rows[t] = series

    # write out csv
    df = pd.DataFrame(rows)
    df.to_csv(f"{filepath}/social_{m}.csv")
    print(f"saved {filepath}/social_{m}.csv")
