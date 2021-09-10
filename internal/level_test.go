package zen_doctor

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLevelConfigs(t *testing.T) {
	for i := Tutorial; i.IsValid(); i++ {
		t.Run(fmt.Sprintf("Level %d", i), func(t *testing.T) {
			level := GetLevel(i)

			// level name is correct
			assert.Equal(t, i, level.Level, "Level index must match")

			// each level must have an updater and win conditions defined
			assert.NotNil(t, level.Updater, "must have updater")
			assert.NotEmpty(t, level.WinConditions, "must have win conditions")
			assert.NotEmpty(t, level.PowerUpLootTable, "must have power ups")
			assert.NotEmpty(t, level.DataLootTable, "must have data")

			// The sum of the loot tables must be 1
			sum := float32(0.0)
			for _, entry := range level.DataLootTable {
				sum += entry.Chance
			}
			assert.Equal(t, float32(1.0), sum, "sum of chances in DataLootTable must equal 1")

			sum = float32(0.0)
			for _, entry := range level.PowerUpLootTable {
				sum += entry.Chance
			}
			assert.Equal(t, float32(1.0), sum, "sum of chances in PowerUpLootTable must equal 1")

			// Loot exists in the loot table for all win conditions
			for _, cond := range level.WinConditions {
				found := false
				for _, loot := range level.DataLootTable {
					if cond.Kind == loot.Data {
						found = true
					}
				}
				assert.True(t, found, "must include loot table for all win conditions")
			}

			// all decay values should be negative
			assert.True(t, 0 > level.LeaveSpeedDecay, "LeaveSpeedDecay must be negative")
			assert.True(t, 0 > level.DataDecayRate, "DataDecayRate must be negative")
			assert.True(t, 0 > level.FootprintDecay, "FootprintDecay must be negative")
			assert.True(t, 0 > level.ThreatDecay, "ThreatDecay must be negative")
			assert.True(t, 0 > level.LootSpeedDecay, "LootSpeedDecay must be negative")
			assert.True(t, 0 > level.PowerUpDecayRate, "LootSpeedDecay must be negative")

			// width and height must be defined
			assert.True(t, level.Width > 0, "Width must be positive")
			assert.True(t, level.Height > 0, "Height must be positive")

			// Data multipliers should be non-zero
			for _, multiplier := range level.DataMultipliers {
				assert.True(t, multiplier > 0, "Data multiplier must be positive")
			}

			// Bit stream chances must be between 0 and 1
			assert.True(t, 0 <= level.BitStreamChance && level.BitStreamChance <= 1, "BitStreamChance must be between 0 and 1")
			assert.True(t, 0 <= level.BadBitChance && level.BadBitChance <= 1, "BadBitChance must be between 0 and 1")
			assert.True(t, 0 <= level.GoodBitChance && level.GoodBitChance <= 1, "GoodBitChance must be between 0 and 1")
		})
	}
}
