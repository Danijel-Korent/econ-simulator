package main

import (
	"flag"
	"fmt"
	"math/rand"
	"os"
	"time"
)

const DEFAULT_CONFIG = "default_configuration.json"

func main() {
	var configFile = flag.String("config", DEFAULT_CONFIG, "specifies the configuration file to use")

	flag.Parse()

	config, err := loadConfig(*configFile)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
	}

	r := rand.New(rand.NewSource(time.Now().UnixMilli()))

	people, producers := initSimulation(config, r)

	detailedMonths := make([]DetailedMonth, config.MaxMonths+1)
	basicMonths := make([]BasicMonthTable, config.MaxMonths+1)
	for month := 0; month < config.MaxMonths; month++ {
		detailedMonths[month] = fillDetailedMonth(people, producers, month)
		basicMonths[month] = fillBasicMonth(people, producers, month)

		producers, people = simulationStep(producers, people, config)

		if month == config.MaxMonths-1 {
			exitSimulation(people, producers, basicMonths, detailedMonths, month)
		}
	}

}

// Steps through one month of the simulation, adjusting variables as needed
func simulationStep(producers []Producer, people []Person, config SimConfig) ([]Producer, []Person) {
	for i := range producers {
		producers[i].simulationStep(producers)
	}

	for i := range people {
		people[i].simulationStep(producers, config)
	}

	return producers, people
}

func initSimulation(config SimConfig, r *rand.Rand) ([]Person, []Producer) {
	producers := make([]Producer, len(config.Producers))
	for i, pConfig := range config.Producers {
		producers[i] = initProducer(config, pConfig, r)
	}

	people := []Person{}
	for i := 0; i < config.NumPeople; i++ {
		person := initPerson(r, i, config)

		producers[person.Employer].Employees = append(producers[person.Employer].Employees, &person)
		people = append(people, person)
	}

	return people, producers
}

func exitSimulation(people []Person, producers []Producer, basicMonths []BasicMonthTable, detailedMonths []DetailedMonth, month int) {
	detailedMonths[month+1] = fillDetailedMonth(people, producers, month)
	basicMonths[month+1] = fillBasicMonth(people, producers, month)
	printSimulationState(basicMonths)

	err := outputSimulationHTML(basicMonths, detailedMonths)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
	}

	fmt.Println("Simulation exiting")
}
