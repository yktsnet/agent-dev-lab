package dbx

import (
	"bytes"
	"fmt"
	"os/exec"
	"strings"
)

// Compose は docker compose exec で mysql クライアントを起動して SQL を流す。
// 1回ごとにプロセスと exec を起動するコストが、元の観測の「場面の数に比例する」待ちの正体なので、
// ドライバで接続を張りっぱなしにする形にはしない。
type Compose struct {
	Project string
	Env     []string
}

func (c Compose) cmd(args ...string) *exec.Cmd {
	base := []string{"compose", "-p", c.Project}
	cmd := exec.Command("docker", append(base, args...)...)
	cmd.Env = append(cmd.Environ(), c.Env...)
	return cmd
}

func (c Compose) Up() error {
	out, err := c.cmd("up", "-d", "--wait").CombinedOutput()
	if err != nil {
		return fmt.Errorf("compose up: %v: %s", err, out)
	}
	return nil
}

func (c Compose) Down() error {
	out, err := c.cmd("down", "-v").CombinedOutput()
	if err != nil {
		return fmt.Errorf("compose down: %v: %s", err, out)
	}
	return nil
}

// Exec は sql を1回の mysql クライアントで流し、出力を返す。
func (c Compose) Exec(db, sql string) (string, error) {
	args := []string{"exec", "-T", "mysql", "mysql", "-uroot", "-plab", "-N", "-B"}
	if db != "" {
		args = append(args, db)
	}
	cmd := c.cmd(args...)
	cmd.Stdin = strings.NewReader(sql)
	var out, stderr bytes.Buffer
	cmd.Stdout, cmd.Stderr = &out, &stderr
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("mysql: %v: %s", err, filterWarn(stderr.String()))
	}
	return strings.TrimSpace(out.String()), nil
}

func filterWarn(s string) string {
	var keep []string
	for _, l := range strings.Split(s, "\n") {
		if !strings.Contains(l, "Using a password") && l != "" {
			keep = append(keep, l)
		}
	}
	return strings.Join(keep, "\n")
}
