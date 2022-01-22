package pkg

type RunOptions struct {
	Endpoint         string
	Snapshot         string
	LogLevel         string
	SkipEtcdRestore  bool
	SkipEtcdStart    bool
	SkipDecodeErrors bool
}
