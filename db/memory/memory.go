// Package memory is a in-memory db.store
package memory

import (
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/asim/nitro/db"
)

// NewStore returns a memory store
func NewStore(opts ...db.Option) db.Store {
	s := &memoryStore{
		options: db.Options{
			Database: "micro",
			Table:    "micro",
		},
		db: make(map[string]*dbValues),
	}
	for _, o := range opts {
		o(&s.options)
	}
	return s
}

type memoryStore struct {
	sync.RWMutex
	options db.Options

	db map[string]*dbValues
}

type dbRecord struct {
	key       string
	value     []byte
	metadata  map[string]interface{}
	expiresAt time.Time
}

type dbValues struct {
	values map[string]*dbRecord
}

func (s *dbValues) Get(key string) (*dbRecord, bool) {
	v, ok := s.values[key]
	if !ok {
		return nil, false
	}
	if v.expiresAt.IsZero() {
		return v, true
	}
	x := v.expiresAt.Sub(time.Now())
	if x.Nanoseconds() > 0 {
		return v, true
	}
	delete(s.values, key)
	return nil, false
}

func (s *dbValues) Delete(key string) {
	delete(s.values, key)
}

func (s *dbValues) Set(key string, v *dbRecord) {
	s.values[key] = v
}

func (s *dbValues) List() map[string]*dbRecord {
	return s.values
}

func (s *dbValues) Flush() {
	s.values = make(map[string]*dbRecord)
}

func (m *memoryStore) prefix(database, table string) string {
	if len(database) == 0 {
		database = m.options.Database
	}
	if len(table) == 0 {
		table = m.options.Table
	}
	return filepath.Join(database, table)
}

func (m *memoryStore) getStore(prefix string) *dbValues {
	m.RLock()
	db := m.db[prefix]
	m.RUnlock()
	if db == nil {
		m.Lock()
		if m.db[prefix] == nil {
			m.db[prefix] = &dbValues{
				values: make(map[string]*dbRecord),
			}
		}
		db = m.db[prefix]
		m.Unlock()
	}
	return db
}

func (m *memoryStore) get(prefix, key string) (*db.Record, error) {
	dbRecord, found := m.getStore(prefix).Get(key)
	if !found {
		return nil, db.ErrNotFound
	}

	// Copy the record on the way out
	newRecord := &db.Record{}
	newRecord.Key = strings.TrimPrefix(dbRecord.key, prefix+"/")
	newRecord.Value = make([]byte, len(dbRecord.value))
	newRecord.Metadata = make(map[string]interface{})

	// copy the value into the new record
	copy(newRecord.Value, dbRecord.value)

	// check if we need to set the expiry
	if !dbRecord.expiresAt.IsZero() {
		newRecord.Expiry = time.Until(dbRecord.expiresAt)
	}

	// copy in the metadata
	for k, v := range dbRecord.metadata {
		newRecord.Metadata[k] = v
	}

	return newRecord, nil
}

func (m *memoryStore) set(prefix string, r *db.Record) {
	// copy the incoming record and then
	// convert the expiry in to a hard timestamp
	i := &dbRecord{}
	i.key = r.Key
	i.value = make([]byte, len(r.Value))
	i.metadata = make(map[string]interface{})

	// copy the the value
	copy(i.value, r.Value)

	// set the expiry
	if r.Expiry != 0 {
		i.expiresAt = time.Now().Add(r.Expiry)
	}

	// set the metadata
	for k, v := range r.Metadata {
		i.metadata[k] = v
	}

	m.getStore(prefix).Set(r.Key, i)
}

func (m *memoryStore) delete(prefix, key string) {
	m.getStore(prefix).Delete(key)
}

func (m *memoryStore) list(prefix string, limit, offset uint, prefixFilter, suffixFilter string) []string {
	allItems := m.getStore(prefix).List()

	allKeys := make([]string, len(allItems))

	// construct list of keys for this prefix
	i := 0
	for k := range allItems {
		allKeys[i] = k
		i++
	}
	keys := make([]string, 0, len(allKeys))
	sort.Slice(allKeys, func(i, j int) bool { return allKeys[i] < allKeys[j] })
	for _, k := range allKeys {
		if prefixFilter != "" && !strings.HasPrefix(k, prefixFilter) {
			continue
		}
		if suffixFilter != "" && !strings.HasSuffix(k, suffixFilter) {
			continue
		}
		if offset > 0 {
			offset--
			continue
		}
		keys = append(keys, k)
		// this check still works if no limit was passed to begin with, you'll just end up with large -ve value
		if limit == 1 {
			break
		}
		limit--
	}
	return keys
}

func (m *memoryStore) Close() error {
	m.Lock()
	defer m.Unlock()
	for _, s := range m.db {
		s.Flush()
	}
	return nil
}

func (m *memoryStore) Init(opts ...db.Option) error {
	for _, o := range opts {
		o(&m.options)
	}
	return nil
}

func (m *memoryStore) String() string {
	return "memory"
}

func (m *memoryStore) Read(key string, opts ...db.ReadOption) ([]*db.Record, error) {
	readOpts := db.ReadOptions{}
	for _, o := range opts {
		o(&readOpts)
	}

	prefix := m.prefix(readOpts.Database, readOpts.Table)

	var keys []string
	// Handle Prefix / suffix
	if readOpts.Prefix || readOpts.Suffix {
		prefixFilter := ""
		if readOpts.Prefix {
			prefixFilter = key
		}
		suffixFilter := ""
		if readOpts.Suffix {
			suffixFilter = key
		}
		keys = m.list(prefix, readOpts.Limit, readOpts.Offset, prefixFilter, suffixFilter)
	} else {
		keys = []string{key}
	}

	var results []*db.Record

	for _, k := range keys {
		r, err := m.get(prefix, k)
		if err != nil {
			return results, err
		}
		results = append(results, r)
	}

	return results, nil
}

func (m *memoryStore) Write(r *db.Record, opts ...db.WriteOption) error {
	writeOpts := db.WriteOptions{}
	for _, o := range opts {
		o(&writeOpts)
	}

	prefix := m.prefix(writeOpts.Database, writeOpts.Table)

	if len(opts) > 0 {
		// Copy the record before applying options, or the incoming record will be mutated
		newRecord := db.Record{}
		newRecord.Key = r.Key
		newRecord.Value = make([]byte, len(r.Value))
		newRecord.Metadata = make(map[string]interface{})
		copy(newRecord.Value, r.Value)
		newRecord.Expiry = r.Expiry

		for k, v := range r.Metadata {
			newRecord.Metadata[k] = v
		}

		m.set(prefix, &newRecord)
		return nil
	}

	// set
	m.set(prefix, r)

	return nil
}

func (m *memoryStore) Delete(key string, opts ...db.DeleteOption) error {
	deleteOptions := db.DeleteOptions{}
	for _, o := range opts {
		o(&deleteOptions)
	}

	prefix := m.prefix(deleteOptions.Database, deleteOptions.Table)
	m.delete(prefix, key)
	return nil
}

func (m *memoryStore) Options() db.Options {
	return m.options
}

func (m *memoryStore) List(opts ...db.ListOption) ([]string, error) {
	listOptions := db.ListOptions{}

	for _, o := range opts {
		o(&listOptions)
	}

	prefix := m.prefix(listOptions.Database, listOptions.Table)
	keys := m.list(prefix, listOptions.Limit, listOptions.Offset, listOptions.Prefix, listOptions.Suffix)
	return keys, nil
}
