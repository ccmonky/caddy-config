package dynconf

import "github.com/ccmonky/typemap"

func init() {
	typemap.MustRegisterType[Callback]()
}
