package conn

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"syscall"

	"polystore_database/src/go/base/id"
)

// uuidlease.go hands out uuid lease numbers (id.Allocator's lease) from a ledger shared by every dataset
// (datastore/shared/uuid_leases.json; not in git), never from a DB. Lease numbers are one sequence across
// datasets, so uuids of different datasets never collide, and the ledger records which dataset took which
// lease. It outlives datastore/runtime/, which is rebuilt per dataset.
//
// Only the Neo4j dump carries uuids (other stores copy them from Neo4j), so the ledger belongs with the dumps:
// a missing ledger is an error, not a silent start at 0; create it from uuid_leases.example.json. Numbers only
// grow and are never reused (after a crash the next run takes a new number).

// UUIDLeaseExampleFile is the ledger template's name (next to the ledger; in git).
const UUIDLeaseExampleFile = "uuid_leases.example.json"

// uuidLedger is the ledger's content.
type uuidLedger struct {
	Comment string       `json:"_comment,omitempty"`
	Next    uint64       `json:"next"` // the next lease number to hand out
	Leases  []leaseEntry `json:"leases"`
}

// leaseEntry records one lease handed out.
type leaseEntry struct {
	Lease   uint64 `json:"lease"`
	Dataset string `json:"dataset"`
}

// LeaseUUIDAllocator takes the next lease number for dataset and returns an allocator for it (for updates;
// call once at process start).
func LeaseUUIDAllocator(ledger, dataset string) (*id.Allocator, error) {
	return leaseUUID(ledger, dataset, false)
}

// LeaseUUIDAllocatorForBulk takes a lease for bulk assignment (to data without uuids). It fails if the dataset
// already took a lease: uuids from it may exist, and reassigning in bulk would break the pairing of the dump
// and the ledger.
func LeaseUUIDAllocatorForBulk(ledger, dataset string) (*id.Allocator, error) {
	return leaseUUID(ledger, dataset, true)
}

// DatasetLeases returns the lease numbers dataset took.
func DatasetLeases(ledger, dataset string) ([]uint64, error) {
	l, err := readLedger(ledger)
	if err != nil {
		return nil, err
	}
	var out []uint64
	for _, e := range l.Leases {
		if e.Dataset == dataset {
			out = append(out, e.Lease)
		}
	}
	return out, nil
}

func readLedger(path string) (*uuidLedger, error) {
	b, err := os.ReadFile(path)
	if errors.Is(err, fs.ErrNotExist) {
		return nil, fmt.Errorf("no uuid lease ledger %s: create it from %s (keep the ledger with the dumps)",
			path, UUIDLeaseExampleFile)
	}
	if err != nil {
		return nil, fmt.Errorf("uuid lease: %w", err)
	}
	var l uuidLedger
	if err := json.Unmarshal(b, &l); err != nil {
		return nil, fmt.Errorf("uuid lease %s: %w", path, err)
	}
	return &l, nil
}

func leaseUUID(path, dataset string, bulk bool) (*id.Allocator, error) {
	if dataset == "" {
		return nil, fmt.Errorf("uuid lease: no dataset")
	}
	// The file is replaced by rename, so lock a separate file (for concurrent processes on one machine).
	lock, err := os.OpenFile(path+".lock", os.O_CREATE|os.O_RDWR, 0o644)
	if err != nil {
		return nil, fmt.Errorf("uuid lease lock: %w", err)
	}
	defer lock.Close()
	if err := syscall.Flock(int(lock.Fd()), syscall.LOCK_EX); err != nil {
		return nil, fmt.Errorf("uuid lease lock: %w", err)
	}
	defer syscall.Flock(int(lock.Fd()), syscall.LOCK_UN)

	l, err := readLedger(path)
	if err != nil {
		return nil, err
	}
	if bulk {
		for _, e := range l.Leases {
			if e.Dataset == dataset {
				return nil, fmt.Errorf("dataset %q already took lease %d (%s); bulk assignment needs a dataset "+
					"without leases. Remove its entries only after reloading its data without uuids", dataset, e.Lease, path)
			}
		}
	}
	a, err := id.NewAllocator(l.Next)
	if err != nil {
		return nil, fmt.Errorf("dataset %q: %w", dataset, err)
	}
	l.Leases = append(l.Leases, leaseEntry{Lease: l.Next, Dataset: dataset})
	l.Next++
	if err := writeFileAtomic(path, l); err != nil {
		return nil, fmt.Errorf("uuid lease: %w", err)
	}
	return a, nil
}

// writeFileAtomic writes a temporary file, fsyncs and renames it (a crash leaves the original intact).
func writeFileAtomic(path string, v any) error {
	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return err
	}
	tmp := path + ".tmp"
	f, err := os.OpenFile(tmp, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o644)
	if err != nil {
		return err
	}
	if _, err := f.Write(append(b, '\n')); err != nil {
		f.Close()
		return err
	}
	if err := f.Sync(); err != nil {
		f.Close()
		return err
	}
	if err := f.Close(); err != nil {
		return err
	}
	if err := os.Rename(tmp, path); err != nil {
		return err
	}
	if d, err := os.Open(filepath.Dir(path)); err == nil {
		_ = d.Sync()
		d.Close()
	}
	return nil
}
