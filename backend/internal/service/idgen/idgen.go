package idgen

import (
	"math/rand"
	"regexp"
	"time"

	"github.com/hasmcp/hasmcp-ce/backend/internal/service/config"
	"github.com/mustafaturan/monoflake"
	zlog "github.com/rs/zerolog/log"
)

type (
	Params struct {
		Config config.Service
	}

	idgenConfig struct {
		Node               uint16 `yaml:"node"`
		EpochTimeInSeconds int64  `yaml:"epochTimeInSeconds"`
		NodeBits           int    `yaml:"nodeBits"`
	}

	Service interface {
		Next() int64
		NextString() string
		ValidStringID(string) bool
	}

	service struct {
		monoflake *monoflake.MonoFlake
	}
)

const (
	_logPrefix = "[idgen] "

	_cfgKey  = "idgen"
	_pattern = "^[0-9a-zA-Z]{11}$"
)

var (
	_regex = regexp.MustCompile(_pattern)
)

// New inits a new id generator
func New(p Params) (Service, error) {
	var cfg idgenConfig
	if err := p.Config.Populate(_cfgKey, &cfg); err != nil {
		return nil, err
	}

	if cfg.Node == 0 {
		cfg.Node = uint16(rand.Intn(1 << 8))
		zlog.Info().Uint16("node", uint16(cfg.Node)).Msg(_logPrefix + "node id is set randomly")
	}

	epoch := time.Unix(cfg.EpochTimeInSeconds, 0)
	f, err := monoflake.New(cfg.Node, monoflake.WithEpoch(epoch), monoflake.WithNodeBits(cfg.NodeBits))
	if err != nil {
		zlog.Error().Str("epoch", epoch.Format(time.RFC3339)).Err(err).Msg(_logPrefix + "failed to init monoflake")
		return nil, err
	}
	zlog.Info().Any("monoflake.cfg", cfg).Msg(_logPrefix + "node is initialized")

	return &service{
		monoflake: f,
	}, nil
}

func (s *service) Next() int64 {
	return s.monoflake.Next().Int64()
}

func (s *service) NextString() string {
	return s.monoflake.Next().String()
}

func (s *service) ValidStringID(id string) bool {
	return _regex.Match([]byte(id))
}
