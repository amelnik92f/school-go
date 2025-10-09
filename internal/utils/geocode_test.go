package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGeocoder_GeocodeAddress(t *testing.T) {
	geocoder := NewGeocoder()

	// Test with a known Berlin address
	address := "Alexanderplatz, 10178 Berlin"
	coords, err := geocoder.GeocodeAddress(address)

	assert.NoError(t, err)
	assert.NotNil(t, coords)

	if coords != nil {
		// Alexanderplatz should be roughly at these coordinates
		assert.InDelta(t, 52.52, coords.Latitude, 0.1, "Latitude should be around 52.52")
		assert.InDelta(t, 13.41, coords.Longitude, 0.1, "Longitude should be around 13.41")
	}
}

func TestGeocoder_GeocodeAddressSafe(t *testing.T) {
	geocoder := NewGeocoder()

	// Test with a valid address
	address := "Brandenburger Tor, 10117 Berlin"
	coords := geocoder.GeocodeAddressSafe(address)

	assert.NotNil(t, coords)

	if coords != nil {
		// Brandenburg Gate should be roughly at these coordinates
		assert.InDelta(t, 52.51, coords.Latitude, 0.1, "Latitude should be around 52.51")
		assert.InDelta(t, 13.37, coords.Longitude, 0.1, "Longitude should be around 13.37")
	}
}

func TestGeocoder_GeocodeAddress_Invalid(t *testing.T) {
	geocoder := NewGeocoder()

	// Test with an invalid address
	address := "This address definitely does not exist xyz123"
	coords, err := geocoder.GeocodeAddress(address)

	assert.Error(t, err)
	assert.Nil(t, coords)
}

func TestGeocoder_GeocodeAddressSafe_Invalid(t *testing.T) {
	geocoder := NewGeocoder()

	// Test with an invalid address - should not panic
	address := "Invalid address xyz123"
	coords := geocoder.GeocodeAddressSafe(address)

	// Should return nil without panicking
	assert.Nil(t, coords)
}
