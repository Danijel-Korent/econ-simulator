package main

//Responsible for the loading, creation, and writing of configuration files

import (
	"encoding/json"
	"fmt"
	"os"
)

// General simulation configuration
type SimConfig struct {
	// Simulation basics
	MaxMonths   int
	PayoutMonth int
	NumPeople   int

	// Individuals
	StartingWalletMin         int
	StartingWalletMax         int
	SavingsRatioMin           float64
	SavingsRatioMax           float64
	FoodIntakeMin             int
	FoodIntakeMax             int
	GasConsumptionPerDistance int
	JobSwitchMultiplier       float64

	// Producers
	Producers []ProducerConfig

	PositionMin int
	PositionMax int
}

// Producer specific options — not all options related to the creation of the producers should necessarily be included in the producers themselves
type ProducerConfig struct {
	ProductName           string
	InitSalary            int
	MaxHires              int
	InitBalance           int
	InitPrice             int
	InitMonthlyProduction int
	InitStock             int
	ProductionLimit       int

	// These variables define the amount above or below one that these variables will change. e.g. a value of 0.05 means that increases will be 1.05 and decreases will be 0.95
	ProductionChangeAmount float64
	PriceChangeAmount      float64

	ProductionCosts []ProductionCost
}

// Loads the given configuration, else loads the default configuration
func loadConfig(configPath string) (SimConfig, error) {
	fmt.Printf("Attempting to load %v \n", configPath)
	_, err := os.Open(configPath)
	if os.IsNotExist(err) {
		if configPath == DEFAULT_CONFIG {
			createDefaultConfigFile(DEFAULT_CONFIG)
		} else {
			panic("Configuration file does not exist")
		}
	}

	fileContents, err := os.ReadFile(configPath)
	if err != nil {
		return SimConfig{}, err
	}

	var config SimConfig
	json.Unmarshal(fileContents, &config)
	fmt.Println("Successfully loaded configuration file")
	return config, nil
}

func createDefaultConfigFile(path string) error {
	fd, err := os.Create(path)

	configBytes, err := marshalDefaultConfig()
	if err != nil {
		return err
	}

	_, err = fd.Write(configBytes)
	if err != nil {
		return err
	}

	return nil
}

// Marshals the default configuration into JSON bytes
func marshalDefaultConfig() ([]byte, error) {
	fmt.Printf("Creating default configuration at %v \n", DEFAULT_CONFIG)
	defaultConfig := getDefaultConfig()

	bytes, err := json.MarshalIndent(defaultConfig, "", "\n")
	if err != nil {
		return nil, err
	}

	return bytes, nil
}

// Outputs a config with sensible defaults, to be used if the config file does not yet exist
func getDefaultConfig() SimConfig {
	defaultProducers := []ProducerConfig{
		{
			ProductName:           "food",
			InitSalary:            10,
			MaxHires:              2,
			InitBalance:           0,
			InitPrice:             10,
			InitMonthlyProduction: 100,
			InitStock:             1000,
			ProductionLimit:       1000,

			ProductionChangeAmount: 0.1,
			PriceChangeAmount:      0.1,

			ProductionCosts: []ProductionCost{{ProducerName: "gasoline", PerUnits: 10, Amount: 1}, {ProducerName: "coffee", PerUnits: 10, Amount: 1}},
		},
		{
			ProductName:           "gasoline",
			InitSalary:            10,
			MaxHires:              2,
			InitBalance:           0,
			InitPrice:             10,
			InitMonthlyProduction: 100,
			InitStock:             1000,
			ProductionLimit:       1000,

			ProductionChangeAmount: 0.1,
			PriceChangeAmount:      0.1,

			ProductionCosts: []ProductionCost{{ProducerName: "gasoline", PerUnits: 10, Amount: 1}, {ProducerName: "coffee", PerUnits: 10, Amount: 1}},
		},
		{
			ProductName:           "coffee",
			InitSalary:            10,
			MaxHires:              2,
			InitBalance:           0,
			InitPrice:             10,
			InitMonthlyProduction: 100,
			InitStock:             1000,
			ProductionLimit:       1000,

			ProductionChangeAmount: 0.1,
			PriceChangeAmount:      0.1,

			ProductionCosts: []ProductionCost{{ProducerName: "gasoline", PerUnits: 10, Amount: 1}, {ProducerName: "coffee", PerUnits: 10, Amount: 1}},
		},
	}

	exampleConfig := SimConfig{
		MaxMonths: 100, PayoutMonth: 49, NumPeople: 20, FoodIntakeMin: 30, FoodIntakeMax: 60, JobSwitchMultiplier: 1.5, PositionMin: 0, PositionMax: 300, GasConsumptionPerDistance: 1, StartingWalletMin: 0, StartingWalletMax: 1000, Producers: defaultProducers, SavingsRatioMin: 0.1, SavingsRatioMax: 0.3,
	}
	return exampleConfig
}
