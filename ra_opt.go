package ra

type RegisterOption func(*registerConfig)

type registerConfig struct {
	global bool
}

func WithGlobal(g bool) RegisterOption {
	return func(c *registerConfig) {
		c.global = g
	}
}

type parseCfg struct {
	ignoreUnknown bool
}

type ParseOpt func(*parseCfg)

func WithIgnoreUnknown(ignore bool) ParseOpt {
	return func(c *parseCfg) {
		c.ignoreUnknown = ignore
	}
}
