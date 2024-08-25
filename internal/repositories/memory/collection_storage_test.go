package memory

import (
	"fmt"
	"math"
	"runtime"
	"testing"

	"github.com/shirou/gopsutil/v3/mem"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUpdate(t *testing.T) {
	collection := NewCollectionMetricStorage()
	initialCount := collection.counter["PollCount"]

	collection.Update()

	if collection.counter["PollCount"] != initialCount+1 {
		t.Errorf("expected PollCount to be %d, got %d", initialCount+1, collection.counter["PollCount"])
	}
}

func AbsPercentageChange[T ~int | ~float64](old, new T) (delta float64) {
	diff := float64(new - old)
	delta = (diff / float64(old)) * 100
	delta = math.Abs(delta)
	return
}

func TestUpdateRuntime(t *testing.T) {
	collection := NewCollectionMetricStorage()

	collection.UpdateRuntime()

	var mem runtime.MemStats
	runtime.ReadMemStats(&mem)

	if AbsPercentageChange(collection.gauge["Alloc"], float64(mem.Alloc)) > 20 {
		t.Errorf("Expected Alloc to be %v, got %v", float64(mem.Alloc), collection.gauge["Alloc"])
	}
}

func TestUpdateGopsutil(t *testing.T) {
	collection := NewCollectionMetricStorage()

	collection.UpdateGopsutil()

	memory, err := mem.VirtualMemory()
	require.NoError(t, err)

	assert.True(
		t,
		AbsPercentageChange(collection.gauge["TotalMemory"], float64(memory.Total)) < 20,
		fmt.Sprintf("Expected TotalMemory to be %v, got %v", float64(memory.Total), collection.gauge["TotalMemory"]),
	)

	assert.True(
		t,
		AbsPercentageChange(collection.gauge["FreeMemory"], float64(memory.Free)) < 20,
		fmt.Sprintf("Expected FreeMemory to be %v, got %v", float64(memory.Free), collection.gauge["FreeMemory"]),
	)

}
