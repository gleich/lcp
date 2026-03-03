package logger

type Operation int

const (
	CacheUpdateMiss Operation = iota
	CacheUpdateHit
	EndpointHit
)

func (o Operation) ToString() string {
	switch o {
	case CacheUpdateMiss:
		return "cache-update-miss"
	case CacheUpdateHit:
		return "cache-update-hit"
	case EndpointHit:
		return "endpoint-hit"
	}
	return ""
}
