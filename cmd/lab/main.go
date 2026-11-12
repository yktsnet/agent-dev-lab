package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/BurntSushi/toml"
	"github.com/yktsnet/agent-dev-lab/internal/dbx"
	"github.com/yktsnet/agent-dev-lab/internal/measure"
	"github.com/yktsnet/agent-dev-lab/internal/scene"
)

type Condition struct {
	Name    string
	Scenes  int
	Collide int
	Prefix  int
}

type Config struct {
	Seed    int64
	Runs    int
	Results string
	DB      struct {
		Project string
		Port    int
		Mycnf   string
	}
	Condition []Condition
}

func main() {
	cfgPath := flag.String("config", "config.toml", "config file")
	flag.Usage = func() {
		fmt.Fprintln(os.Stderr, "usage: lab [-config file] seed-all|down")
		flag.PrintDefaults()
	}
	flag.Parse()
	var cfg Config
	if _, err := toml.DecodeFile(*cfgPath, &cfg); err != nil {
		fatal(err)
	}
	c := dbx.Compose{Project: cfg.DB.Project, Env: []string{
		fmt.Sprintf("LAB_PORT=%d", cfg.DB.Port), "LAB_MYCNF=" + cfg.DB.Mycnf,
	}}
	switch flag.Arg(0) {
	case "seed-all":
		fatal(seedAll(cfg, c))
	case "down":
		fatal(c.Down())
	default:
		flag.Usage()
		os.Exit(2)
	}
}

func seedAll(cfg Config, c dbx.Compose) error {
	if err := c.Up(); err != nil {
		return err
	}
	w, err := measure.Open(cfg.Results)
	if err != nil {
		return err
	}
	defer w.Close()
	for _, cond := range cfg.Condition {
		scenes := scene.Generate(cfg.Seed, cond.Scenes, scene.Overlap{Collide: cond.Collide, Prefix: cond.Prefix})
		for run := 1; run <= cfg.Runs; run++ {
			fmt.Printf("%s run %d/%d\n", cond.Name, run, cfg.Runs)
			stages, err := seedOnce(c, scenes)
			if err != nil {
				return fmt.Errorf("%s run %d: %w", cond.Name, run, err)
			}
			for _, s := range stages {
				w.Add(measure.Row{Condition: cond.Name, Run: run, Lane: 1, Stage: s.name, Seconds: s.sec})
			}
		}
	}
	return nil
}

type stage struct {
	name string
	sec  float64
}

// seedOnce は DB を作り直し、土台と全場面を投入して expect で確かめる。
func seedOnce(c dbx.Compose, scenes []scene.Scene) ([]stage, error) {
	var out []stage
	step := func(name string, f func() error) error {
		sec, err := measure.Time(f)
		out = append(out, stage{name, sec})
		return err
	}
	if err := step("reset", func() error { return reset(c) }); err != nil {
		return nil, err
	}
	if err := step("base", func() error { _, err := c.Exec("lab", scene.Base()); return err }); err != nil {
		return nil, err
	}
	if err := step("seed", func() error {
		for _, s := range scenes {
			if _, err := c.Exec("lab", s.Seed); err != nil {
				return fmt.Errorf("seed %s: %w", s.Name, err)
			}
		}
		return nil
	}); err != nil {
		return nil, err
	}
	if err := step("expect", func() error {
		for _, s := range scenes {
			for _, ch := range s.Expect {
				got, err := c.Exec("lab", ch.SQL)
				if err != nil {
					return fmt.Errorf("expect %s: %w", s.Name, err)
				}
				if got != ch.Want {
					return fmt.Errorf("expect %s: %q: got %q, want %q", s.Name, ch.SQL, got, ch.Want)
				}
			}
		}
		return nil
	}); err != nil {
		return nil, err
	}
	return out, nil
}

func reset(c dbx.Compose) error {
	if _, err := c.Exec("", "DROP DATABASE IF EXISTS lab; CREATE DATABASE lab;"); err != nil {
		return err
	}
	files, err := filepath.Glob("migrations/*.sql")
	if err != nil {
		return err
	}
	sort.Strings(files)
	for _, f := range files {
		b, err := os.ReadFile(f)
		if err != nil {
			return err
		}
		if _, err := c.Exec("lab", string(b)); err != nil {
			return fmt.Errorf("%s: %w", f, err)
		}
	}
	return nil
}

func fatal(err error) {
	if err != nil {
		fmt.Fprintln(os.Stderr, "error:", strings.TrimSpace(err.Error()))
		os.Exit(1)
	}
}
