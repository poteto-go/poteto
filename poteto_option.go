package poteto

type PotetoOption struct {
	WithRequestId   bool   `yaml:"with_request_id" env:"WITH_REQUEST_ID" envDefault:"true"`
	DebugMode       bool   `yaml:"debug_mode" env:"DEBUG_MODE" envDefault:"false"`
	ListenerNetwork string `yaml:"listener_network" env:"LISTENER_NETWORK" envDefault:"tcp"`
}

var DefaultPotetoOption = PotetoOption{}
