package errors

// SafeAttributer is an attributer that only returns attributes that are safe.
type SafeAttributer struct {
	Attributer

	// SafeAttributes are the attributes that are safe to return.
	SafeAttributes []string
}

// Attributes returns the safe attributes.
func (i *SafeAttributer) Attributes() Attributes {
	attrs := i.Attributer.Attributes()

	res := make(Attributes, len(i.SafeAttributes))

	for _, key := range i.SafeAttributes {
		if value, ok := attrs[key]; ok {
			res[key] = value
		}
	}

	return res
}

// Safe takes an attribute and a list of safe attribute names and returns a SafeAttributer.
func Safe(a Attributer, whitelist []string) *SafeAttributer {
	return &SafeAttributer{
		Attributer:     a,
		SafeAttributes: whitelist,
	}
}
