// Copyright © 2018 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package store

// NewResultFunc represents a constructor of some arbitrary type.
type NewResultFunc func() interface{}

// Client represents a generic interface to interact with different store implementations in CRUD manner.
//
// Create creates a new PrimaryKey, stores v under that key and returns it.
// Find searches for the value associated with PrimaryKey specified and stores it in v. v must be a pointer type.
// FindBy returns mapping of PrimaryKey -> value, which match field values specified in filter. Filter represents an AND relation,
// meaning that only entries matching all the fields in filter should be returned.
// newResultFunc is the constructor of a single value expected to be returned.
// Update calculates the diff between old and new values and overwrites stored fields under PrimaryKey with that.
// Delete deletes the value stored under PrimaryKey specified.
type Client interface {
	Create(v interface{}) (PrimaryKey, error)
	Find(id PrimaryKey, v interface{}) error
	FindBy(filter interface{}, newResult NewResultFunc) (map[PrimaryKey]interface{}, error)
	Update(id PrimaryKey, new, old interface{}) error
	Delete(id PrimaryKey) error
}

type typedStoreClient struct {
	TypedStore
}

// NewTypedStoreClient returns a new instance of the Client, which uses TypedStore as the storing backend.
func NewTypedStoreClient(s TypedStore) Client {
	return &typedStoreClient{s}
}

func (cl *typedStoreClient) Create(v interface{}) (PrimaryKey, error) {
	return cl.TypedStore.Create(MarshalMap(v))
}

func (cl *typedStoreClient) Find(id PrimaryKey, v interface{}) error {
	m, err := cl.TypedStore.Find(id)
	if err != nil {
		return err
	}
	return UnmarshalMap(m, v)
}

func (cl *typedStoreClient) FindBy(filter interface{}, newResult NewResultFunc) (map[PrimaryKey]interface{}, error) {
	m, err := cl.TypedStore.FindBy(MarshalMap(filter))
	if err != nil {
		return nil, err
	}

	filtered := make(map[PrimaryKey]interface{}, len(m))
	for k, v := range m {
		iface := newResult()
		if err = UnmarshalMap(v, iface); err != nil {
			return nil, err
		}
		filtered[k] = iface
	}
	return filtered, nil
}

func (cl *typedStoreClient) Update(id PrimaryKey, new, old interface{}) error {
	diff := Diff(MarshalMap(new), MarshalMap(old))
	if len(diff) == 0 {
		return nil
	}
	return cl.TypedStore.Update(id, diff)
}

type byteStoreClient struct {
	ByteStore
}

// NewByteStoreClient returns a new instance of the Client, which uses ByteStore as the storing backend.
func NewByteStoreClient(s ByteStore) Client {
	return &byteStoreClient{s}
}

func (cl *byteStoreClient) Create(v interface{}) (PrimaryKey, error) {
	m, err := MarshalByteMap(v)
	if err != nil {
		return nil, err
	}
	return cl.ByteStore.Create(m)
}

func (cl *byteStoreClient) Find(id PrimaryKey, v interface{}) error {
	m, err := cl.ByteStore.Find(id)
	if err != nil {
		return err
	}
	return UnmarshalByteMap(m, v)
}

func (cl *byteStoreClient) FindBy(filter interface{}, newResult NewResultFunc) (map[PrimaryKey]interface{}, error) {
	fm, err := MarshalByteMap(filter)
	if err != nil {
		return nil, err
	}

	m, err := cl.ByteStore.FindBy(fm)
	if err != nil {
		return nil, err
	}
	filtered := make(map[PrimaryKey]interface{}, len(m))
	for k, v := range m {
		iface := newResult()
		if err = UnmarshalByteMap(v, iface); err != nil {
			return nil, err
		}
		filtered[k] = iface
	}
	return filtered, nil
}

func (cl *byteStoreClient) Update(id PrimaryKey, new, old interface{}) error {
	nm, err := MarshalByteMap(new)
	if err != nil {
		return err
	}
	om, err := MarshalByteMap(old)
	if err != nil {
		return err
	}
	diff := ByteDiff(nm, om)
	if len(diff) == 0 {
		return nil
	}
	return cl.ByteStore.Update(id, diff)
}
