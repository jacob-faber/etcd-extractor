package pkg

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"github.com/openshift/api"
	"go.etcd.io/etcd/clientv3"
	"go.uber.org/zap"
	jsonserializer "k8s.io/apimachinery/pkg/runtime/serializer/json"
	"k8s.io/kubectl/pkg/scheme"
	"os"
	"strconv"
	"strings"
	"time"
)

var (
	DefaultEtcdTimeout = 15 * time.Second
)

func init() {
	// Register OpenShift and external k8s.io/api types, so we will be able to decode them from protobuf to YAML
	api.Install(scheme.Scheme)
	api.InstallKube(scheme.Scheme)
}

type EtcdService struct {
	logger *zap.Logger
	opts   *RunOptions

	client *clientv3.Client

	pidFile string
}

func (s *EtcdService) Connect(ctx context.Context) error {
	var tlsConfig *tls.Config

	config := clientv3.Config{
		Endpoints:   []string{s.opts.Endpoint},
		Context:     ctx,
		TLS:         tlsConfig,
		DialTimeout: time.Second,
	}
	client, err := clientv3.New(config)
	if err != nil {
		return fmt.Errorf("unable to connect to etcd: %v\n", err)
	}
	s.client = client
	return nil
}

func (s *EtcdService) Close() {
	_ = s.client.Close()
}

func (s *EtcdService) PrintKeysWithPrefix(ctx context.Context, key string) error {
	var resp *clientv3.GetResponse

	resp, err := s.client.Get(ctx, key, clientv3.WithPrefix(), clientv3.WithKeysOnly())
	if err != nil {
		return err
	}

	var sb strings.Builder
	for _, kv := range resp.Kvs {
		sb.WriteString(string(kv.Key) + "\n")
	}
	fmt.Print(sb.String())

	return nil
}

func (s *EtcdService) PrintAllKeys(ctx context.Context) error {
	var resp *clientv3.GetResponse

	// Gets all keys with starting ord(key) >= ord("/"), all ASCII numbers and characters are above the character "/"
	resp, err := s.client.Get(ctx, "/", clientv3.WithFromKey(), clientv3.WithKeysOnly())
	if err != nil {
		return err
	}

	var sb strings.Builder
	for _, kv := range resp.Kvs {
		sb.WriteString(string(kv.Key) + "\n")
	}
	fmt.Print(sb.String())

	return nil
}

func (s *EtcdService) RestoreSnapshot(ctx context.Context) error {
	if s.opts.SkipEtcdRestore || alreadyRunning(s.pidFile) {
		s.logger.Debug("skipping restoring snapshot. It's been already done.")
		return nil
	}

	err := IsFile(s.opts.Snapshot)
	if err != nil {
		return err
	}

	s.logger.Debug("Restoring etcd...")
	err = RunCmd(ctx, s.logger, "etcdctl", "snapshot", "restore", s.opts.Snapshot)
	if err != nil {
		return err
	}
	s.logger.Debug("Restored successfully")

	return nil
}

func (s *EtcdService) PrintValues(ctx context.Context, keys []string, skipErrors bool) error {
	for i, key := range keys {
		resp, err := s.client.Get(ctx, key)
		// Not found value for the key
		if len(resp.Kvs) == 0 && err == nil {
			// Let's try to use the key as a prefix
			resp, err = s.client.Get(ctx, key, clientv3.WithPrefix())
		}
		if len(resp.Kvs) == 0 && err == nil {
			if !skipErrors {
				s.logger.Warn("key " + key + " not found")
			}
			continue
		}
		if err != nil {
			return err
		}

		decoder := scheme.Codecs.UniversalDeserializer()
		encoder := jsonserializer.NewSerializerWithOptions(
			jsonserializer.DefaultMetaFactory,
			scheme.Scheme,
			scheme.Scheme,
			jsonserializer.SerializerOptions{Yaml: true},
		)

		sb := new(strings.Builder)
		for j, kv := range resp.Kvs {
			obj, _, err := decoder.Decode(kv.Value, nil, nil)
			if err != nil {
				if skipErrors {
					continue
				}
				return fmt.Errorf("unable to decode %s: %v\n. Use flag --skip-errors to silently skip unknown resources", kv.Key, err)
			}

			err = encoder.Encode(obj, sb)
			if err != nil {
				if skipErrors {
					continue
				}
				return fmt.Errorf("unable to encode %s: %v\n. Use flag --skip-errors to silently skip unknown resources", kv.Key, err)
			}

			// Add YAML separator if not last
			if j != len(resp.Kvs)-1 {
				sb.Write([]byte("---\n"))
			}
		}

		// Add YAML separator if not last
		if i != len(keys)-1 {
			sb.Write([]byte("---\n"))
		}

		fmt.Print(sb.String())
	}
	return nil
}

func (s *EtcdService) StartServer(ctx context.Context) error {
	if s.opts.SkipEtcdStart || alreadyRunning(s.pidFile) {
		s.logger.Debug("skipping starting etcd server")
		return nil
	}

	err := writePidFile(s.pidFile)
	if err != nil {
		return err
	}

	s.logger.Debug("Starting etcd...")
	err = RunCmd(ctx, s.logger,
		"etcd",
		"--auto-compaction-retention", "0",
		"--name", "default",
		"--listen-client-urls", "http://0.0.0.0:2379",
		"--advertise-client-urls", "http://0.0.0.0:2379",
		"--listen-peer-urls", "http://0.0.0.0:2380",
	)
	if err != nil {
		return err
	}

	err = deletePidFile(s.pidFile)
	if err != nil {
		return err
	}

	return nil
}

func (s *EtcdService) WaitForReady(ctx context.Context) error {
	ticker := time.NewTicker(time.Second)
	select {
	case <-ctx.Done():
		ticker.Stop()
		return ctx.Err()
	case <-ticker.C:
		resp, err := s.client.Status(ctx, s.opts.Endpoint)
		if err != nil {
			return err
		}
		if resp.Version != "" {
			ticker.Stop()
			return nil
		}
	}
	return nil
}

func NewEtcdService(logger *zap.Logger, opts *RunOptions) *EtcdService {
	return &EtcdService{
		logger:  logger,
		opts:    opts,
		pidFile: "/tmp/etcd-extractor.pid",
	}
}

func alreadyRunning(pidFilePath string) bool {
	if err := IsFile(pidFilePath); errors.Is(err, os.ErrNotExist) {
		return false
	} else {
		return true
	}
}

func writePidFile(pidFilePath string) error {
	pid := strconv.Itoa(os.Getpid())
	return os.WriteFile(pidFilePath, []byte(pid), 0644)
}

func deletePidFile(pidFilePath string) error {
	return os.Remove(pidFilePath)
}
