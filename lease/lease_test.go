package lease

import "testing"

func TestGrantSetsExpiry(t *testing.T) {
	st := &State{}
	got := Grant(st, "L1", "h1", 0, 3000, &Counter{})
	if got.Status != "granted" || got.ExpiryMS != 3000 {
		t.Fatalf("授予后到期时刻应是 3000，得到 %+v", got)
	}
}

func TestGrantSameIDKeepsOneLease(t *testing.T) {
	st := &State{}
	Grant(st, "L1", "h1", 0, 3000, &Counter{})
	Grant(st, "L1", "h2", 0, 3000, &Counter{})
	if len(st.Leases) != 1 {
		t.Fatalf("同 id 重复授予应只有一份，得到 %d", len(st.Leases))
	}
}

func TestRenewInsideWindow(t *testing.T) {
	st := &State{}
	Grant(st, "L1", "h1", 0, 3000, &Counter{})
	got := Renew(st, "L1", "r1", 2500, &Counter{})
	if got.Status != "ok" || got.ExpiryMS != 5500 || got.Renewals != 1 {
		t.Fatalf("窗口内续约应成功并顺延，得到 %+v", got)
	}
}

func TestRenewUnknownLease(t *testing.T) {
	st := &State{}
	if got := Renew(st, "nope", "r1", 0, &Counter{}); got.Status != "unknown" {
		t.Fatalf("未知租约应报 unknown，得到 %+v", got)
	}
}

func TestStatusUnknownLease(t *testing.T) {
	st := &State{}
	if got := Status(st, "nope", 0, &Counter{}); got.Status != "unknown" {
		t.Fatalf("未知租约应报 unknown，得到 %+v", got)
	}
}
