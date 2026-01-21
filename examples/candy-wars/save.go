package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

const (
	maxSaveSlots = 5
	saveVersion  = 1
)

// SaveData contains all game state for serialization.
type SaveData struct {
	Version  int       `json:"version"`
	SavedAt  time.Time `json:"saved_at"`
	SlotName string    `json:"slot_name"`

	// Core state
	Cash      int `json:"cash"`
	Debt      int `json:"debt"`
	Bank      int `json:"bank"`
	BankLimit int `json:"bank_limit"`
	Day       int `json:"day"`
	Hour      int `json:"hour"`
	Location  int `json:"location"`
	Heat      int `json:"heat"`
	HP        int `json:"hp"`

	// Inventory
	Inventory Inventory `json:"inventory"`
	Stash     Inventory `json:"stash"`
	Capacity  int       `json:"capacity"`

	// Upgrades
	WorkoutCount       int  `json:"workout_count"`
	ThickSkinCount     int  `json:"thick_skin_count"`
	TrackPracticeCount int  `json:"track_practice_count"`
	HasMuscle          bool `json:"has_muscle"`
	HasIntimidation    bool `json:"has_intimidation"`
	BackpackTier       int  `json:"backpack_tier"`
	HasStash           bool `json:"has_stash"`
	StashCapacity      int  `json:"stash_capacity"`
	HasBike            bool `json:"has_bike"`
	HasInformant       bool `json:"has_informant"`
	BankExpanded       bool `json:"bank_expanded"`

	// Equipment
	OwnedWeapons   []string `json:"owned_weapons"`
	OwnedArmors    []string `json:"owned_armors"`
	EquippedWeapon string   `json:"equipped_weapon"`
	EquippedArmor  string   `json:"equipped_armor"`

	// Difficulty
	Difficulty string `json:"difficulty"`
}

// SaveSlotInfo contains metadata for save slot display.
type SaveSlotInfo struct {
	Slot     int
	Name     string
	Day      int
	NetWorth int
	SavedAt  time.Time
	Empty    bool
}

func getSaveDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".candy-wars", "saves"), nil
}

func getSavePath(slot int) (string, error) {
	dir, err := getSaveDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, fmt.Sprintf("slot-%d.json", slot)), nil
}

// ListSaveSlots returns info about all save slots.
func ListSaveSlots() ([]SaveSlotInfo, error) {
	slots := make([]SaveSlotInfo, maxSaveSlots)

	for i := 0; i < maxSaveSlots; i++ {
		slots[i] = SaveSlotInfo{Slot: i + 1, Empty: true}

		path, err := getSavePath(i + 1)
		if err != nil {
			continue
		}

		data, err := os.ReadFile(path)
		if err != nil {
			continue
		}

		var save SaveData
		if err := json.Unmarshal(data, &save); err != nil {
			continue
		}

		slots[i] = SaveSlotInfo{
			Slot:     i + 1,
			Name:     save.SlotName,
			Day:      save.Day,
			NetWorth: save.Cash + save.Bank - save.Debt,
			SavedAt:  save.SavedAt,
			Empty:    false,
		}
	}

	return slots, nil
}

// SaveToSlot saves the game to a slot.
func (g *Game) SaveToSlot(slot int, name string) error {
	if slot < 1 || slot > maxSaveSlots {
		return fmt.Errorf("invalid slot: %d", slot)
	}

	dir, err := getSaveDir()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}

	data := g.toSaveData()
	data.SlotName = name
	data.SavedAt = time.Now()
	data.Version = saveVersion

	bytes, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return err
	}

	path, err := getSavePath(slot)
	if err != nil {
		return err
	}
	return os.WriteFile(path, bytes, 0o644)
}

// LoadFromSlot loads a game from a slot.
func (g *Game) LoadFromSlot(slot int) error {
	if slot < 1 || slot > maxSaveSlots {
		return fmt.Errorf("invalid slot: %d", slot)
	}

	path, err := getSavePath(slot)
	if err != nil {
		return err
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	var save SaveData
	if err := json.Unmarshal(data, &save); err != nil {
		return err
	}

	g.fromSaveData(&save)
	return nil
}

// DeleteSaveSlot removes a save slot.
func DeleteSaveSlot(slot int) error {
	path, err := getSavePath(slot)
	if err != nil {
		return err
	}
	return os.Remove(path)
}

func (g *Game) toSaveData() SaveData {
	ownedWeapons := make([]string, 0)
	for w := range g.ownedWeapons {
		ownedWeapons = append(ownedWeapons, w)
	}
	ownedArmors := make([]string, 0)
	for a := range g.ownedArmors {
		ownedArmors = append(ownedArmors, a)
	}

	difficulty := string(g.difficulty)
	if difficulty == "" {
		difficulty = string(DifficultyNormal)
	}

	return SaveData{
		Cash:               g.Cash.Get(),
		Debt:               g.Debt.Get(),
		Bank:               g.Bank.Get(),
		BankLimit:          g.BankLimit,
		Day:                g.Day.Get(),
		Hour:               g.Hour.Get(),
		Location:           g.Location.Get(),
		Heat:               g.Heat.Get(),
		HP:                 g.HP.Get(),
		Inventory:          g.Inventory.Get(),
		Stash:              g.stash,
		Capacity:           g.Capacity,
		WorkoutCount:       g.workoutCount,
		ThickSkinCount:     g.thickSkinCount,
		TrackPracticeCount: g.trackPracticeCount,
		HasMuscle:          g.hasMuscle,
		HasIntimidation:    g.hasIntimidation,
		BackpackTier:       g.backpackTier,
		HasStash:           g.hasStash,
		StashCapacity:      g.stashCapacity,
		HasBike:            g.hasBike,
		HasInformant:       g.hasInformant,
		BankExpanded:       g.bankExpanded,
		OwnedWeapons:       ownedWeapons,
		OwnedArmors:        ownedArmors,
		EquippedWeapon:     g.equippedWeapon,
		EquippedArmor:      g.equippedArmor,
		Difficulty:         difficulty,
	}
}

func (g *Game) fromSaveData(save *SaveData) {
	g.GameOver.Set(false)
	g.GameOverMsg.Set("")
	g.ShowEvent.Set(false)
	g.EventTitle.Set("")
	g.EventMessage.Set("")
	g.Combat = nil
	g.eventOptions = nil
	g.Loans = nil

	g.Cash.Set(save.Cash)
	g.Debt.Set(save.Debt)
	g.Bank.Set(save.Bank)
	g.BankLimit = save.BankLimit
	g.Day.Set(save.Day)
	g.Hour.Set(save.Hour)
	g.Location.Set(save.Location)
	g.Heat.Set(save.Heat)
	g.HP.Set(save.HP)
	g.Inventory.Set(save.Inventory)
	g.stash = save.Stash
	g.Capacity = save.Capacity
	g.workoutCount = save.WorkoutCount
	g.thickSkinCount = save.ThickSkinCount
	g.trackPracticeCount = save.TrackPracticeCount
	g.hasMuscle = save.HasMuscle
	g.hasIntimidation = save.HasIntimidation
	g.backpackTier = save.BackpackTier
	g.hasStash = save.HasStash
	g.stashCapacity = save.StashCapacity
	g.hasBike = save.HasBike
	g.hasInformant = save.HasInformant
	g.bankExpanded = save.BankExpanded

	g.ownedWeapons = make(map[string]bool)
	for _, w := range save.OwnedWeapons {
		g.ownedWeapons[w] = true
	}
	g.ownedArmors = make(map[string]bool)
	for _, a := range save.OwnedArmors {
		g.ownedArmors[a] = true
	}
	g.equippedWeapon = save.EquippedWeapon
	g.equippedArmor = save.EquippedArmor

	g.difficulty = Difficulty(save.Difficulty)
	if g.difficulty == "" {
		g.difficulty = DifficultyNormal
	}
	if _, ok := DifficultySettings[g.difficulty]; !ok {
		g.difficulty = DifficultyNormal
	}

	g.RefreshPrices()
}
