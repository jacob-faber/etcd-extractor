FROM quay.io/coreos/etcd:v3.5.3

COPY ./bin/etcd-extractor.amd64 /usr/local/bin/etcd-extractor

ENTRYPOINT ["/usr/local/bin/etcd-extractor"]
CMD ["help"]
