package main

import (
	"encoding/json"
	"fmt"
	"math"
	"math/rand"
	"os"
	"time"
)

const CONFIG_FILE_NAME = "config.json"

type ProducerConfig struct {
	ProductName           string
	InitSalary            int
	MaxHires              int
	InitBalance           int
	InitPrice             int
	InitMonthlyProduction int
	InitStock             int

	//These variables define the amount above or below one that these variables will change. e.g. a value of 0.05 means that increases will be 1.05 and decreases will be 0.95
	ProductionChangeAmount float64
	PriceChangeAmount      float64

	ProductionCosts []ProductionCost
}

type SimConfig struct {
	//Simulation basics
	MaxMonths   int
	PayoutMonth int
	NumPeople   int

	//Individuals
	StartingWalletMin         int
	StartingWalletMax         int
	FoodIntakeMin             int
	FoodIntakeMax             int
	GasConsumptionPerDistance int
	JobSwitchMultiplier       float64

	//Producers
	Producers []ProducerConfig

	PositionMin int
	PositionMax int
}

type ProductionCost struct {
	ProducerName string
	PerUnits     int
	Amount       int
}

type Producer struct {
	BankBalance  int
	Product      string
	MonthSalary  int
	MonthHires   int
	Employees    []*Person
	NumEmployees int
	Price        int
	Stock        int
	//Number of units since the producer last bought necessary materials (gas and coffee)
	UnpaidUnits       int
	MonthlyProduction int
	PosX              int
	PosY              int

	MaxHires               int
	ProductionChangeAmount float64
	PriceChangeAmount      float64

	ProductionCosts []ProductionCost
}

type Person struct {
	IdNumber          int
	Employer          int
	WalletAmount      int
	Salary            int
	MonthlyFoodIntake int
	MonthlyGasIntake  int
	PosX              int
	PosY              int
}

// Simulates the purchase of goods, adjusting variables on the person and alerting the producer
func (p *Person) buyGoods(producers []Producer) {
	foodProducerIdx := findProducerIdx("food", producers)
	gasProducerIdx := findProducerIdx("gasoline", producers)
	coffeeProducerIdx := findProducerIdx("coffee", producers)
	foodCost := producers[foodProducerIdx].registerPurchase(p.getUnitsToPurchase(producers[foodProducerIdx], p.MonthlyFoodIntake))
	p.WalletAmount -= foodCost
	gasCost := producers[gasProducerIdx].registerPurchase(p.getUnitsToPurchase(producers[gasProducerIdx], p.MonthlyGasIntake))

	p.WalletAmount -= gasCost
	if p.WalletAmount > foodCost {
		p.WalletAmount -= producers[foodProducerIdx].registerPurchase(p.MonthlyFoodIntake)
	}

	maxCoffee := producers[coffeeProducerIdx].getMaxUnits(p.WalletAmount)
	p.WalletAmount -= producers[coffeeProducerIdx].registerPurchase(maxCoffee)
}

// Returns either the desired number of units to purchase by the individual or the maximum amount they can purchase with their wallet amount
func (p *Person) getUnitsToPurchase(producer Producer, desiredIntake int) int {
	if p.WalletAmount <= 0 {
		return 0
	}

	if p.WalletAmount >= producer.Price*desiredIntake {
		return desiredIntake
	} else {
		return producer.getMaxUnits(p.WalletAmount)
	}
}

// Get paid by the employer
func (p *Person) receiveSalary(producers []Producer) {
	p.Salary = producers[p.Employer].MonthSalary
	p.WalletAmount += p.Salary
}

// Look for a new job at a producer if the salary is JOB_SWITCH_MULTIPLIER higher
func (p *Person) checkNewJobs(producers []Producer, config SimConfig) {
	for i, producer := range producers {
		if float64(producer.MonthSalary)/float64(p.Salary) >= config.JobSwitchMultiplier && i != p.Employer {
			if producers[i].addEmployee(p) {
				producers[p.Employer].removeEmployee(p)
				p.Employer = i
				return
			}
		}
	}
}

func (p *Person) calculateGasConsumption(producers []Producer, config SimConfig) {
	employer := producers[p.Employer]
	p.MonthlyGasIntake = pythagDistance(p.PosX, p.PosY, employer.PosX, employer.PosY) * config.GasConsumptionPerDistance
}

// Adjusts the price and salary of employees based on the stock
func (p *Producer) adjustVariables() {
	newPrice := 0.0
	newProduction := 0.0
	if p.Stock == 0 {
		newProduction = float64(p.MonthlyProduction) * (1.0 + p.ProductionChangeAmount)
		newPrice = float64(p.Price) * (1.0 + p.PriceChangeAmount)
	} else {
		newProduction = float64(p.MonthlyProduction) * (1.0 - p.ProductionChangeAmount)
		newPrice = float64(p.Price) * (1.0 - p.PriceChangeAmount)
	}
	p.MonthlyProduction = int(newProduction + 0.5)
	p.Price = int(newPrice + 0.5)
}

// Adds as much product to the producer as they have money to make
func (p *Producer) produceProducts() {
	p.Stock += p.MonthlyProduction
	p.UnpaidUnits += p.MonthlyProduction
}

func (p *Producer) payProductionCost(producers []Producer) {
	for _, cost := range p.ProductionCosts {
		desiredPurchases := int(float64(p.UnpaidUnits) / float64(cost.PerUnits))
		producerIdx := findProducerIdx(cost.ProducerName, producers)
		purchasableUnits := producers[producerIdx].getMaxUnits(p.BankBalance)
		units := desiredPurchases
		if desiredPurchases < purchasableUnits {
			units = purchasableUnits
		}

		p.BankBalance -= producers[producerIdx].registerPurchase(units)
		p.UnpaidUnits -= units * cost.PerUnits
	}
}

// Removes the given employee from the producer
func (p *Producer) removeEmployee(person *Person) {
	for i := range p.Employees {
		if p.Employees[i].IdNumber == person.IdNumber {
			p.Employees = append(p.Employees[:i], p.Employees[i+1:]...)
			p.NumEmployees -= 1
			return
		}
	}
}

func (p *Producer) calculateSalary() {
	if p.NumEmployees > 0 {
		p.MonthSalary = p.BankBalance / p.NumEmployees
		return
	}

	p.MonthSalary = p.BankBalance
}

// Checks if a new employee can be hired. Employs them and returns true if so, returns false otherwise.
func (p *Producer) addEmployee(person *Person) bool {
	if p.MonthHires < p.MaxHires {
		p.Employees = append(p.Employees, person)
		p.NumEmployees += 1
		p.MonthHires += 1
		return true
	}

	return false
}

// Subtracts from stock, adding to bank balance. Returns the cost of purchase.
func (p *Producer) registerPurchase(amount int) int {
	if p.Stock >= amount {
		p.Stock -= amount
		p.BankBalance += amount * p.Price
		return amount * p.Price
	} else {
		price := p.Stock * p.Price
		p.Stock = 0
		return price
	}
}

// Returns the maximum number of units one can buy with a certain amount of money
func (p *Producer) getMaxUnits(money int) int {
	//Cast instead of rounding to truncate (prevents overspends)
	amount := int(float64(money) / float64(p.Price))
	if amount > p.Stock {
		return p.Stock
	}
	return amount
}

func findProducerIdx(name string, producers []Producer) int {
	for i, p := range producers {
		if p.Product == name {
			return i
		}
	}
	return -1
}

func main() {
	config, err := loadConfig()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
	}

	r := rand.New(rand.NewSource(time.Now().UnixMilli()))

	producers := make([]Producer, len(config.Producers))
	for i, pConfig := range config.Producers {
		producers[i] = initProducer(config, pConfig, r)
	}

	people := []Person{}
	for i := 0; i < config.NumPeople; i++ {
		person := initPerson(r, i, config)

		producers[person.Employer].NumEmployees += 1
		producers[person.Employer].Employees = append(producers[person.Employer].Employees, &person)
		people = append(people, person)

	}

	detailedMonths := make([]DetailedMonth, config.MaxMonths+1)
	basicMonths := make([]BasicMonthTable, config.MaxMonths+1)
	for month := 0; month < config.MaxMonths; month++ {
		detailedMonth := fillDetailedMonth(people, producers, month)
		basicMonth := fillBasicMonth(people, producers, month)
		detailedMonths[month] = detailedMonth
		basicMonths[month] = basicMonth

		producers, people = simulationStep(producers, people, month, config)

		if month == config.MaxMonths-1 {
			detailedMonth := fillDetailedMonth(people, producers, month)
			basicMonth := fillBasicMonth(people, producers, month)
			detailedMonths[month+1] = detailedMonth
			basicMonths[month+1] = basicMonth

			printSimulationState(basicMonths)

			err := outputSimulationHTML(basicMonths, detailedMonths)
			if err != nil {
				fmt.Fprintln(os.Stderr, err)
			}
		}
	}

	fmt.Println("Simulation exiting")
}

func loadConfig() (SimConfig, error) {
	_, err := os.Open(CONFIG_FILE_NAME)
	if os.IsNotExist(err) {
		err := createConfigIfNotExists()
		if err != nil {
			return SimConfig{}, err
		}
	}

	fileContents, err := os.ReadFile(CONFIG_FILE_NAME)
	if err != nil {
		return SimConfig{}, err
	}

	var config SimConfig
	json.Unmarshal(fileContents, &config)

	return config, nil
}

// Steps through one month of the simulation, adjusting variables as needed
func simulationStep(producers []Producer, people []Person, month int, config SimConfig) ([]Producer, []Person) {
	for i := range producers {
		producers[i].MonthHires = 0
		producers[i].adjustVariables()
		producers[i].payProductionCost(producers)
		producers[i].calculateSalary()
		producers[i].produceProducts()
	}

	for i := range people {
		people[i].checkNewJobs(producers, config)
		people[i].calculateGasConsumption(producers, config)
		people[i].buyGoods(producers)
		people[i].receiveSalary(producers)
		if month == config.PayoutMonth {
			people[i].WalletAmount *= 2
		}
	}

	return producers, people
}

// Initialises a new producer and returns it
func initProducer(config SimConfig, pConfig ProducerConfig, r *rand.Rand) Producer {

	return Producer{BankBalance: pConfig.InitBalance, Product: pConfig.ProductName, Price: pConfig.InitPrice, MaxHires: pConfig.MaxHires, Stock: pConfig.InitStock, MonthSalary: pConfig.InitSalary, Employees: []*Person{}, NumEmployees: 0, MonthlyProduction: pConfig.InitMonthlyProduction, PosX: randIntInRange(config.PositionMin, config.PositionMax, r), PosY: randIntInRange(config.PositionMin, config.PositionMax, r), ProductionChangeAmount: pConfig.ProductionChangeAmount, PriceChangeAmount: pConfig.PriceChangeAmount, ProductionCosts: pConfig.ProductionCosts, UnpaidUnits: 0}
}

// Creates a new person, generating random variables. Returns the person and the producer they are employed by
func initPerson(r *rand.Rand, ID int, config SimConfig) Person {
	randomEmployer := randIntInRange(0, len(config.Producers), r)

	return Person{
		IdNumber:          ID,
		Employer:          randomEmployer,
		WalletAmount:      randIntInRange(config.StartingWalletMin, config.StartingWalletMax, r),
		Salary:            0,
		MonthlyFoodIntake: randIntInRange(config.FoodIntakeMin, config.FoodIntakeMax, r),
		PosX:              randIntInRange(config.PositionMin, config.PositionMax, r),
		PosY:              randIntInRange(config.PositionMin, config.PositionMax, r),
	}
}

// Generates a random integer between min and max
func randIntInRange(min int, max int, r *rand.Rand) int {
	return r.Intn(max-min) + min
}

// Debug function for testing how money enters the simulation
func calculateTotalMoneyInSimulation(people []Person, producers []Producer) int {
	total := 0
	for _, p := range people {
		total += p.WalletAmount
	}
	for _, p := range producers {
		total += p.BankBalance
	}

	return total
}

// Outputs a config with sensible defaults, to be used if the config file does not yet exist
func createConfigIfNotExists() error {
	file, err := os.Create(CONFIG_FILE_NAME)
	if err != nil {
		return err
	}

	defer file.Close()

	defaultProducers := []ProducerConfig{
		{
			ProductName:           "food",
			InitSalary:            10,
			MaxHires:              2,
			InitBalance:           0,
			InitPrice:             10,
			InitMonthlyProduction: 100,
			InitStock:             1000,

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

			ProductionChangeAmount: 0.1,
			PriceChangeAmount:      0.1,

			ProductionCosts: []ProductionCost{{ProducerName: "gasoline", PerUnits: 10, Amount: 1}, {ProducerName: "coffee", PerUnits: 10, Amount: 1}},
		},
	}

	exampleConfig := SimConfig{
		MaxMonths: 100, PayoutMonth: 49, NumPeople: 20, FoodIntakeMin: 30, FoodIntakeMax: 60, JobSwitchMultiplier: 1.5, PositionMin: 0, PositionMax: 300, GasConsumptionPerDistance: 1, StartingWalletMin: 0, StartingWalletMax: 1000, Producers: defaultProducers,
	}

	bytes, err := json.MarshalIndent(exampleConfig, "", "\t")
	if err != nil {
		return err
	}

	_, err = file.Write(bytes)
	if err != nil {
		return err
	}

	err = file.Sync()
	if err != nil {
		return err
	}

	return nil
}

// Calculates the straight line distance between two points
func pythagDistance(x1 int, y1 int, x2 int, y2 int) int {
	distanceX := math.Abs(float64(x2 - x1))
	distanceY := math.Abs(float64(y2 - y1))
	return int(math.Sqrt(math.Pow(distanceX, 2) + math.Pow(distanceY, 2)))
}
