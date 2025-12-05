package disk

import (
	"fmt"
	"os"

	"github.com/hawkneo/utils/web/health"
	"golang.org/x/sys/unix"
)

const GB = 1024 * 1024 * 1024

var (
	_ health.Indicator = (*diskIndicator)(nil)
)

type diskIndicator struct {
	threshold float64
}

// NewDiskIndicator creates a new disk indicator with the given threshold.
//
// threshold is the threshold of disk usage rate.
func NewDiskIndicator(threshold float64) health.Indicator {
	if threshold <= 0 || threshold > 1 {
		panic("threshold must be in (0, 1]")
	}

	return &diskIndicator{
		threshold: threshold,
	}
}

func (d diskIndicator) Name() string {
	return "Disk"
}

func (d diskIndicator) Health() health.Health {
	wd, err := os.Getwd()
	if err != nil {
		return health.NewUnknownHealth(err)
	}

	var stat unix.Statfs_t
	if err := unix.Statfs(wd, &stat); err != nil {
		return health.NewUnknownHealth(err)
	}

	allSpace := stat.Blocks * uint64(stat.Bsize)
	freeSpace := stat.Bfree * uint64(stat.Bsize)
	usedSpace := allSpace - freeSpace

	h := health.NewUpHealth()
	if float64(usedSpace)/float64(allSpace) > d.threshold {
		h.Status = health.Down
	}

	h.Details["all"] = fmt.Sprintf("%.2f GB", float64(allSpace)/GB)
	h.Details["used"] = fmt.Sprintf("%.2f GB", float64(usedSpace)/GB)
	h.Details["free"] = fmt.Sprintf("%.2f GB", float64(freeSpace)/GB)
	return h
}
