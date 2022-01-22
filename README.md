# Etcd Extractor

Extracts resources from OpenShift/Kubernetes ETCD backup.

## Use cases

* Restoring something deleted
* Safely inspecting etcd database (especially with **wait** command) from snapshot instead of real system

### Usage

```
Usage:
  etcd-extractor [command]

Available Commands:
  completion  Generate the autocompletion script for the specified shell
  get         Get all values (or subset if specifying stdin, args or files as args as keys) in YAML format in etcd database
  help        Help about any command
  list        List all keys in etcd database based on prefix (default is "/")
  version     Print version
  wait        Restore etcd database and start etcd server running infinitely

Flags:
      --endpoint string     URL pointing to running etcd instance (default "http://127.0.0.1:2379")
  -h, --help                help for etcd-extractor
      --loglevel string     Specify log level (debug, info, warn, error, dpanic, panic, fatal) (default "info")
      --skip-etcd-restore   Skip restoring etcd
      --skip-etcd-start     Skip starting etcd
      --snapshot string     File path to the snapshot location (default "/tmp/snapshot.db")

Use "etcd-extractor [command] --help" for more information about a command.
```

## Most performant usage

#### Terminal 1

```shell
docker run -it --rm --name etcd-extractor \
  -v "$PWD/snapshot_2022-01-21_060859.db:/tmp/snapshot.db:ro" \
  camabeh/etcd-extractor:latest wait
```

#### Terminal 2 - it will be much faster as the instance is already running and is ready

```shell
docker exec -it etcd-extractor etcd-extractor list

# get with ARGS
docker exec -it etcd-extractor etcd-extractor get /kubernetes.io/configmaps/openshift-logging/fluentd /kubernetes.io/clusterrolebindings/system:basic-user
# get with STDIN
cat keys.txt | docker exec -i etcd-extractor etcd-extractor get
```


## Less performant usage

### List all keys in etcd snapshot

```shell
docker run -it --rm \
  -v "$PWD/snapshot_2022-01-21_060859.db:/tmp/snapshot.db:ro" \
  camabeh/etcd-extractor:latest list
```

### Get all values for keys (can be passed as STDIN or ARGS)

```shell
docker run -it --rm \
  -v "$PWD/snapshot_2022-01-21_060859.db:/tmp/snapshot.db:ro" \
  camabeh/etcd-extractor:latest get /kubernetes.io/configmaps/openshift-logging/fluentd /kubernetes.io/clusterrolebindings/system:basic-user
```

## Raw etcd introspection

```shell
# Restores DB and starts etcd server and waits infinitely
docker run -it --rm --name etcd-extractor \
  -v "$PWD/snapshot_2022-01-21_060859.db:/tmp/snapshot.db:ro" \
  camabeh/etcd-extractor:latest get /kubernetes.io/secrets/openshift-logging

docker exec -it etcd-extractor sh
# Gets all keys with / prefix
$ etcdctl get / --prefix --keys-only | head
# Shows the value, but encoded in protocol buffer format
$ etcdctl get /kubernetes.io/configmaps/openshift-logging/fluentd
```

---

Heavily inspired by [etcdhelper](https://github.com/openshift/origin/blob/master/tools/etcdhelper/etcdhelper.go).
