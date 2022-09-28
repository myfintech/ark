package utils

import (
	"testing"

	"github.com/brianvoe/gofakeit/v4"
	"github.com/tj/assert"
)

type Location struct {
	Country string `fake:"{address.country}"`
	City    string `fake:"{address.city}"`
	Person  *Person
}

type Person struct {
	FirstName string `fake:"{person.first}"`
	LastName  string `fake:"{person.last}"`
}

func TestBuildDeepMapString(t *testing.T) {
	t.Run("should be able to build an arbitrarily deep map[string]interface{}", func(t *testing.T) {
		addressBook := make(map[string]interface{})
		locations := make([]*Location, 10)

		for i := range locations {
			person := &Person{}
			location := &Location{
				Person: person,
			}
			locations[i] = location
			gofakeit.Struct(location)
			gofakeit.Struct(person)
		}

		country := gofakeit.Address().Country
		for _, location := range locations {
			keys := []string{country, location.City}
			BuildDeepMapString(location.Person, keys, addressBook)
		}
		// FIXME: improve this test to not just validate that an interface has multiple keys. This is weak.
		assert.True(t, len(addressBook[country].(map[string]interface{})) > 1, "should create deeply nested values and not overwrite the map")
	})

	t.Run("should be okay with an empty keys slice", func(t *testing.T) {
		testData := map[string]interface{}{}
		BuildDeepMapString(1, []string{}, testData)
		assert.Empty(t, testData)
	})

}
