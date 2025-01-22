package kadmin

import (
	"ktea/sradmin"
)

type SraSetter interface {
	SetSra(sra sradmin.SrAdmin)
}

func (ka *SaramaKafkaAdmin) SetSra(sra sradmin.SrAdmin) {
	ka.sra = sra
}
