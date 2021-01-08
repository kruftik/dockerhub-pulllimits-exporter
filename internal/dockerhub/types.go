package dockerhub

import (
	"sync"
	"time"
)

type Limit struct {
	Limit           int
	RefreshDuration time.Duration
}

type LimitsList struct {
	Total     Limit
	Remaining Limit
}

type LimitsRetriever struct {
	m sync.Mutex

	image struct {
		Label string
		Tag   string
	}

	AuthProvider *authProviderT

	limits LimitsList
}
