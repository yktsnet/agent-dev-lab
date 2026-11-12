package scene

import (
	"fmt"
	"math/rand"
	"strings"
)

type Check struct {
	SQL  string
	Want string
}

type Scene struct {
	Name   string
	ID     string
	Seed   string
	Expect []Check
}

// Overlap は場面どうしの ID の重なり方を決める。
type Overlap struct {
	// Collide 組の場面が同じ event ID を持つ。全場面の投入が主キーの重複で落ちる
	Collide int
	// Prefix 組で、片方の expect が ID の先頭の文字列で行を選び、もう片方の ID がその文字列で始まる
	Prefix int
}

const baseVenues = 4
const baseAttendees = 20

// Base は全場面が前提にする共通の行を返す。
func Base() string {
	var b strings.Builder
	for i := 1; i <= baseVenues; i++ {
		fmt.Fprintf(&b, "INSERT INTO venues VALUES ('VN-%02d','Venue %d',%d);\n", i, i, 100*i)
	}
	tiers := []string{"standard", "member", "staff"}
	for i := 1; i <= baseAttendees; i++ {
		fmt.Fprintf(&b, "INSERT INTO attendees VALUES ('AT-%03d','Attendee %d','%s');\n", i, i, tiers[i%len(tiers)])
	}
	return b.String()
}

// Generate は seed から決定的に n 個の場面を作る。
// 場面どうしは土台の行（会場・参加者）を共有して参照するので、文字列としては ID が重なって見えるが衝突ではない。
func Generate(seed int64, n int, ov Overlap) []Scene {
	r := rand.New(rand.NewSource(seed))
	ids := make([]string, n+1)
	prefixVictim := map[int]bool{}
	for i := 1; i <= n; i++ {
		ids[i] = fmt.Sprintf("EV-%04d", i)
	}
	// 末尾の場面から順に、前半の場面と1組ずつ重ねる。前半の相手は組ごとに別にする
	victims := r.Perm(n / 2)
	k := n
	for c := 0; c < ov.Collide && k > n/2; c, k = c+1, k-1 {
		ids[k] = ids[1+victims[0]]
		victims = victims[1:]
	}
	for p := 0; p < ov.Prefix && k > n/2; p, k = p+1, k-1 {
		v := 1 + victims[0]
		victims = victims[1:]
		prefixVictim[v] = true
		ids[k] = ids[v] + "X"
	}
	out := make([]Scene, 0, n)
	for i := 1; i <= n; i++ {
		id := ids[i]
		status := []string{"open", "closed", "draft"}[r.Intn(3)]
		venue := fmt.Sprintf("VN-%02d", 1+r.Intn(baseVenues))
		nres := 2 + r.Intn(7)
		var b strings.Builder
		fmt.Fprintf(&b, "INSERT INTO events VALUES ('%s','%s','Scene %d','%s');\n", id, venue, i, status)
		seats := 0
		for j := 1; j <= nres; j++ {
			s := 1 + r.Intn(4)
			seats += s
			fmt.Fprintf(&b, "INSERT INTO reservations VALUES ('%s-R%02d','%s','AT-%03d',%d,'confirmed');\n",
				id, j, id, 1+r.Intn(baseAttendees), s)
		}
		count := fmt.Sprintf("SELECT COUNT(*) FROM reservations WHERE event_id='%s'", id)
		if prefixVictim[i] {
			count = fmt.Sprintf("SELECT COUNT(*) FROM reservations WHERE id LIKE '%s%%'", id)
		}
		out = append(out, Scene{
			Name: fmt.Sprintf("scene-%04d", i),
			ID:   id,
			Seed: b.String(),
			Expect: []Check{
				{SQL: fmt.Sprintf("SELECT status FROM events WHERE id='%s'", id), Want: status},
				{SQL: count, Want: fmt.Sprint(nres)},
				{SQL: fmt.Sprintf("SELECT SUM(seats) FROM reservations WHERE event_id='%s'", id), Want: fmt.Sprint(seats)},
			},
		})
	}
	return out
}
