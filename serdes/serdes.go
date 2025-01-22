package serdes

type Deserializer interface {
	Deserialize(data []byte) (string, error)
}
