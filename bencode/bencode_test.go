package bencode

import (
	"fmt"
	"log"
	"testing"
)

func TestBencode(t *testing.T) {
	// Example struct for marshaling
	type Example struct {
		Name  string         `bencode:"name"`
		Age   int            `bencode:"age"`
		Tags  []string       `bencode:"tags"`
		Attrs map[string]int `bencode:"attrs"`
	}

	// Create an example instance
	example := Example{
		Name:  "John Doe",
		Age:   30,
		Tags:  []string{"developer", "golang"},
		Attrs: map[string]int{"height": 180, "weight": 75},
	}

	serializer := NewBencodeSerializer()
	// Marshal the example struct
	data, err := serializer.Marshal(example)
	if err != nil {
		log.Fatalf("Marshal error: %v", err)
	}
	fmt.Printf("Marshalled data: %s\n", string(data))

	// Unmarshal the data back into a struct
	var decoded Example
	decoded.Attrs = map[string]int{}

	err = serializer.Unmarshal(data, &decoded)
	if err != nil {
		log.Fatalf("Unmarshal error: %v", err)
	}
	fmt.Printf("Unmarshalled struct: %+v\n", decoded)
}
