FROM victoriametrics/victoria-metrics

COPY targets.yaml /targets.yaml

CMD ["-promscrape.config=/targets.yaml", "-search.latencyOffset=0s"]
