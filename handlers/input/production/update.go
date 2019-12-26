package production

import (
	"time"

	models "github.com/theMomax/openefs/models/production"
	"github.com/theMomax/openefs/utils/metadata"
)

type update struct {
	data *models.Data
	time time.Time
	meta metadata.Metadata
}

func (u *update) Data() *models.Data {
	return u.data
}

func (u *update) Time() time.Time {
	return u.time
}

func (u *update) Meta() metadata.Metadata {
	return u.meta
}
