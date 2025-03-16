package addrs

type AddressMap map[string]interface{}

func (m AddressMap) Add(key string, value interface{}) bool {
	if _, exists := m[key]; exists {
		return true
	}
	m[key] = value
	return false
}
