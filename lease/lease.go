// Package lease 管理分布式租约的授予与续约。
package lease

// 续约参数。
const (
	GraceMS     = 1000
	MaxRenewals = 3
)

// Lease 是一份租约。
type Lease struct {
	ID       string `json:"id"`
	Holder   string `json:"holder"`
	ExpiryMS int    `json:"expiry_ms"`
	TTLMS    int    `json:"ttl_ms"`
	Renewals int    `json:"renewals"`
}

// State 是落盘状态。
type State struct {
	Leases []Lease  `json:"leases"`
	Seen   []string `json:"seen"`

	index map[string]int
	seen  map[string]struct{}
}

// Counter 统计本轮检查过的租约条数。
type Counter struct {
	Scanned int
}

// Result 是一次操作的输出。
type Result struct {
	ID       string `json:"id"`
	Status   string `json:"status"`
	ExpiryMS int    `json:"expiry_ms"`
	Renewals int    `json:"renewals"`
	Scanned  int    `json:"scanned"`
}

func find(s *State, id string, c *Counter) *Lease {
	if s.index == nil {
		s.index = make(map[string]int, len(s.Leases))
		for i := range s.Leases {
			c.Scanned++
			if _, ok := s.index[s.Leases[i].ID]; !ok {
				s.index[s.Leases[i].ID] = i
			}
		}
	}
	c.Scanned++
	i, ok := s.index[id]
	if !ok {
		return nil
	}
	return &s.Leases[i]
}

func (s *State) seenReq(reqID string) bool {
	if s.seen == nil {
		s.seen = make(map[string]struct{}, len(s.Seen))
		for _, r := range s.Seen {
			s.seen[r] = struct{}{}
		}
	}
	_, ok := s.seen[reqID]
	return ok
}

// Grant 授予一份租约；同一个 id 重复授予时保留原来那份。
func Grant(s *State, id, holder string, atMS, ttlMS int, c *Counter) Result {
	if existing := find(s, id, c); existing != nil {
		return Result{ID: id, Status: "exists", ExpiryMS: existing.ExpiryMS,
			Renewals: existing.Renewals, Scanned: c.Scanned}
	}
	s.Leases = append(s.Leases, Lease{ID: id, Holder: holder, ExpiryMS: atMS + ttlMS, TTLMS: ttlMS})
	s.index[id] = len(s.Leases) - 1
	return Result{ID: id, Status: "granted", ExpiryMS: atMS + ttlMS, Renewals: 0, Scanned: c.Scanned}
}

// Renew 续约一份租约。
func Renew(s *State, id, reqID string, atMS int, c *Counter) Result {
	l := find(s, id, c)
	if l == nil {
		return Result{ID: id, Status: "unknown", Scanned: c.Scanned}
	}
	if s.seenReq(reqID) {
		return Result{ID: id, Status: "replay", ExpiryMS: l.ExpiryMS,
			Renewals: l.Renewals, Scanned: c.Scanned}
	}
	switch {
	case atMS >= l.ExpiryMS:
		return Result{ID: id, Status: "expired", ExpiryMS: l.ExpiryMS,
			Renewals: l.Renewals, Scanned: c.Scanned}
	case atMS < l.ExpiryMS-GraceMS:
		return Result{ID: id, Status: "too-early", ExpiryMS: l.ExpiryMS,
			Renewals: l.Renewals, Scanned: c.Scanned}
	case l.Renewals >= MaxRenewals:
		return Result{ID: id, Status: "limit", ExpiryMS: l.ExpiryMS,
			Renewals: l.Renewals, Scanned: c.Scanned}
	}
	l.ExpiryMS = atMS + l.TTLMS
	l.Renewals++
	s.Seen = append(s.Seen, reqID)
	s.seen[reqID] = struct{}{}
	return Result{ID: id, Status: "ok", ExpiryMS: l.ExpiryMS, Renewals: l.Renewals, Scanned: c.Scanned}
}

// Status 查一份租约当前的样子。
func Status(s *State, id string, atMS int, c *Counter) Result {
	l := find(s, id, c)
	if l == nil {
		return Result{ID: id, Status: "unknown", Scanned: c.Scanned}
	}
	status := "active"
	if atMS >= l.ExpiryMS {
		status = "expired"
	}
	return Result{ID: id, Status: status, ExpiryMS: l.ExpiryMS, Renewals: l.Renewals, Scanned: c.Scanned}
}
