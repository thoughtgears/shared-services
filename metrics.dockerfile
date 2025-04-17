FROM us-docker.pkg.dev/cloud-ops-agents-artifacts/google-cloud-opentelemetry-collector/otelcol-google:0.121.0

COPY metrics.yaml /etc/otelcol-google/config.yaml