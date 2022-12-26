package dynconf

// Not implemnt...
type Nacos struct{}

func (n Nacos) ID() string {
	return "config.ext.dynconf.nacos"
}
