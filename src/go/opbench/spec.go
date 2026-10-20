// Package opbench は論理演算子を 1 つずつ、条件を振って測る実験基盤。
//
// 実験は JSON の定義ファイル（Spec）に書き、`go run ./cmd/polystore -mode opbench -spec <file>` で実行する。
// 各実験は対象の論理演算子 1 つと、その入力を作る最小の前段だけからなる論理計画を直接組み立て
// （`Argument`＝ ldbc_sf1 から抜き出した N 件の uuid → 前段 → 対象）、配置・入力件数・反復モデル・チャンク幅などを
// 掃引して実行する。計画の根の物理ステップを「対象」として、その時間と全体の時間・往復数・行数を CSV に記録し、
// 結果が graph 配置と一致するかも確かめる。データは読むだけ（非 graph 配置は移行済みの列を使う）。
//
// 定義ファイルの書き方は opbench/specs/README.md、例は opbench/specs/*.json。
package opbench

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"time"
)

// Spec は実験の定義ファイル。
type Spec struct {
	Name        string       `json:"name"`
	Warmup      int          `json:"warmup"`  // 計測しない事前実行の回数（条件ごと）
	Repeat      int          `json:"repeat"`  // 計測する回数（条件ごと）
	Timeout     string       `json:"timeout"` // 1 回の実行の上限（例 "60s"。既定 60s）
	Experiments []Experiment `json:"experiments"`
}

// Experiment は対象の論理演算子 1 つの実験。使うフィールドは Op による（specs/README.md の表）。
type Experiment struct {
	Name string `json:"name"`
	Op   string `json:"op"` // EntityScan / GetProperty / Expand / VarLengthExpand / Filter / Aggregation / Sort / Limit / Distinct

	Label  string `json:"label"`  // 入力の頂点のラベル（uuid を抜き出す・走査する）
	Prop   string `json:"prop"`   // 対象のプロパティ（GetProperty / Filter / Sort / Distinct / Aggregation のグループキー）
	Sample string `json:"sample"` // 入力を抜き出す条件（Cypher の WHERE。変数は n）。空なら全件から

	Rel string `json:"rel"` // Expand / VarLengthExpand: 関係タイプ
	Dir string `json:"dir"` // out / in / both（既定 out）
	To  string `json:"to"`  // 行き先のラベル（空なら制約なし）
	Min int    `json:"min"` // VarLengthExpand のホップ数
	Max int    `json:"max"`

	Cmp   string `json:"cmp"`   // Filter / EntityScan: 比較演算子（= <> < > <= >= IN）
	Value any    `json:"value"` // 比較の値（IN なら配列）

	Agg     string `json:"agg"`     // Aggregation: count / sum / avg / min / max / collect
	AggProp string `json:"aggProp"` // 集約の引数のプロパティ（空なら count(*)）
	Desc    bool   `json:"desc"`    // Sort
	Count   int    `json:"count"`   // Limit

	Sweep Sweep `json:"sweep"`
}

// Sweep は掃引する軸。空の軸は既定の 1 点（括弧内）になる。
type Sweep struct {
	Placement   []string `json:"placement"`   // 対象のプロパティを置くストア（graph）
	Rows        []int    `json:"rows"`        // 入力の件数（1000）。input=scan では無視
	Model       []string `json:"model"`       // 反復モデル（bulk）
	Chunk       []int    `json:"chunk"`       // settings.MaterializeChunkSize（今の値。0 はストア既定）
	Parallelism []int    `json:"parallelism"` // settings.StoreFetchParallelism（今の値）
	Vector      []int    `json:"vector"`      // settings.VectorSize（今の値）
	At          []string `json:"at"`          // store / coordinator: 対象をストアに任せるか、コーディネータで実行させるか（store）
	Input       []string `json:"input"`       // ids（抜き出した uuid）/ scan（ラベルの全件走査。吸収できる演算子の比較用）（ids）
}

// Condition は掃引の 1 点。
type Condition struct {
	Placement   string
	Rows        int
	Model       string
	Chunk       int // -1 は今の設定のまま
	Parallelism int // 0 は今の設定のまま
	Vector      int // 0 は今の設定のまま
	At          string
	Input       string
}

// Load は定義ファイルを読んで確かめる。
func Load(path string) (*Spec, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var s Spec
	dec := json.NewDecoder(bytes.NewReader(data))
	dec.DisallowUnknownFields() // 綴り違いのフィールドを黙って無視しない
	if err := dec.Decode(&s); err != nil {
		return nil, fmt.Errorf("opbench: %s: %w", path, err)
	}
	if err := s.Validate(); err != nil {
		return nil, fmt.Errorf("opbench: %s: %w", path, err)
	}
	return &s, nil
}

// Validate は定義の矛盾を確かめ、既定値を埋める。
func (s *Spec) Validate() error {
	if s.Name == "" {
		return fmt.Errorf("name が空")
	}
	if s.Repeat <= 0 {
		s.Repeat = 3
	}
	if s.Timeout == "" {
		s.Timeout = "60s"
	}
	if _, err := time.ParseDuration(s.Timeout); err != nil {
		return fmt.Errorf("timeout: %w", err)
	}
	if len(s.Experiments) == 0 {
		return fmt.Errorf("experiments が空")
	}
	for i := range s.Experiments {
		e := &s.Experiments[i]
		if e.Name == "" {
			e.Name = fmt.Sprintf("%s#%d", e.Op, i+1)
		}
		if err := e.validate(); err != nil {
			return fmt.Errorf("experiment %q: %w", e.Name, err)
		}
	}
	return nil
}

func (e *Experiment) validate() error {
	need := func(field, v string) error {
		if v == "" {
			return fmt.Errorf("%s には %s が要る", e.Op, field)
		}
		return nil
	}
	if err := need("label", e.Label); err != nil {
		return err
	}
	switch e.Op {
	case "EntityScan":
		if e.Cmp != "" {
			if err := need("prop", e.Prop); err != nil {
				return err
			}
		}
	case "GetProperty", "Sort", "Distinct":
		if err := need("prop", e.Prop); err != nil {
			return err
		}
	case "Filter":
		if err := need("prop", e.Prop); err != nil {
			return err
		}
		if err := need("cmp", e.Cmp); err != nil {
			return err
		}
	case "Expand", "VarLengthExpand":
		if err := need("rel", e.Rel); err != nil {
			return err
		}
		switch e.Dir {
		case "", "out", "in", "both":
		default:
			return fmt.Errorf("dir は out / in / both: %q", e.Dir)
		}
		if e.Op == "VarLengthExpand" && (e.Min < 0 || (e.Max >= 0 && e.Max < e.Min)) {
			return fmt.Errorf("ホップ数 min=%d max=%d", e.Min, e.Max)
		}
	case "Aggregation":
		if err := need("agg", e.Agg); err != nil {
			return err
		}
	case "Limit":
		if e.Count <= 0 {
			return fmt.Errorf("Limit には count（正）が要る")
		}
	default:
		return fmt.Errorf("未対応の op %q", e.Op)
	}
	for _, at := range e.Sweep.At {
		if at != "store" && at != "coordinator" {
			return fmt.Errorf("at は store / coordinator: %q", at)
		}
		if at == "coordinator" && !barrierable(e.Op) {
			return fmt.Errorf("%s は at=coordinator を選べない（常にストアで実行する演算子）", e.Op)
		}
	}
	for _, in := range e.Sweep.Input {
		if in != "ids" && in != "scan" {
			return fmt.Errorf("input は ids / scan: %q", in)
		}
	}
	if e.Op == "EntityScan" && len(e.Sweep.Input) > 0 {
		return fmt.Errorf("EntityScan は常に走査（input は指定しない）")
	}
	for _, r := range e.Sweep.Rows {
		if r <= 0 {
			return fmt.Errorf("rows は正: %d", r)
		}
	}
	return nil
}

// barrierable はコーディネータで実行させられる（at=coordinator を選べる）演算子か。
func barrierable(op string) bool {
	switch op {
	case "Filter", "Aggregation", "Sort", "Limit", "Distinct":
		return true
	}
	return false
}

// Conditions は掃引の全点（軸の直積）を決まった順で返す。
func (e *Experiment) Conditions() []Condition {
	strs := func(xs []string, def string) []string {
		if len(xs) == 0 {
			return []string{def}
		}
		return xs
	}
	ints := func(xs []int, def int) []int {
		if len(xs) == 0 {
			return []int{def}
		}
		return xs
	}
	inputs := strs(e.Sweep.Input, "ids")
	if e.Op == "EntityScan" {
		inputs = []string{"scan"}
	}
	var out []Condition
	for _, in := range inputs {
		rows := ints(e.Sweep.Rows, 1000)
		if in == "scan" {
			rows = []int{0} // 走査は件数を振らない
		}
		for _, r := range rows {
			for _, pl := range strs(e.Sweep.Placement, "graph") {
				for _, at := range strs(e.Sweep.At, "store") {
					for _, m := range strs(e.Sweep.Model, "bulk") {
						for _, c := range ints(e.Sweep.Chunk, -1) {
							for _, p := range ints(e.Sweep.Parallelism, 0) {
								for _, v := range ints(e.Sweep.Vector, 0) {
									out = append(out, Condition{Placement: pl, Rows: r, Model: m, Chunk: c,
										Parallelism: p, Vector: v, At: at, Input: in})
								}
							}
						}
					}
				}
			}
		}
	}
	return out
}

// MaxRows は抜き出す入力の件数（掃引の最大。小さい件数はその先頭を使うので、件数どうしで入力が入れ子になる）。
func (e *Experiment) MaxRows() int {
	max := 0
	for _, c := range e.Conditions() {
		if c.Rows > max {
			max = c.Rows
		}
	}
	return max
}
