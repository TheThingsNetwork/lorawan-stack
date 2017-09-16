// Copyright © 2017 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package store

type NewResultFunc func() interface{}

// Client represents a generic interface to interact with different store implementations
type Client interface {
	Create(v interface{}) (PrimaryKey, error)
	Find(id PrimaryKey, v interface{}) error
	FindBy(filter interface{}, newResult func() interface{}) (map[PrimaryKey]interface{}, error)
	Update(id PrimaryKey, new, old interface{}) error
	Delete(id PrimaryKey) error
}

type typedStoreClient struct {
	TypedStore
}

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

func (cl *typedStoreClient) FindBy(filter interface{}, newResult func() interface{}) (map[PrimaryKey]interface{}, error) {
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

func (cl *byteStoreClient) FindBy(filter interface{}, newResult func() interface{}) (map[PrimaryKey]interface{}, error) {
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
