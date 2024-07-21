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

type SimConfig struct {
	//Simulation basics
	MaxMonths   int
	PayoutMonth int
	NumPeople   int

	//Individuals
	FoodIntakeMin             int
	FoodIntakeMax             int
	GasConsumptionPerDistance int
	JobSwitchMultiplier       float64

	//Producers
	InitSalary               int
	MaxHires                 int
	InitBalance              int
	InitPrice                int
	InitMonthlyProduction    int
	InitStock                int
	ProductionUnitCostAmount int
	ProductionCoffeeCost     int
	ProductionGasCost        int
	InitWithStock            bool

	PositionMin int
	PositionMax int
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
	foodCost := producers[FoodIdx].registerPurchase(p.getUnitsToPurchase(producers[FoodIdx], p.MonthlyFoodIntake))
	p.WalletAmount -= foodCost
	gasCost := producers[GasolineIdx].registerPurchase(p.getUnitsToPurchase(producers[GasolineIdx], p.MonthlyGasIntake))
	p.WalletAmount -= gasCost
	if p.WalletAmount > foodCost {
		fmt.Println("Buying additional food")
		p.WalletAmount -= producers[FoodIdx].registerPurchase(p.MonthlyFoodIntake)
	}

	maxCoffee := producers[CoffeeIdx].getMaxUnits(p.WalletAmount)
	p.WalletAmount -= producers[CoffeeIdx].registerPurchase(maxCoffee)
}

// Returns either the desired number of units to purchase by the individual or the maximum amount they can purchase with their wallet amount
func (p *Person) getUnitsToPurchase(producer Producer, desiredIntake int) int {
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

// Look for a new job at a producer if the  salary is JOB_SWITCH_MULTIPLIER higher
func (p *Person) checkNewJobs(producers []Producer, config SimConfig) {
	for i, producer := range producers {
		if float64(producer.MonthSalary)/float64(p.Salary) >= config.JobSwitchMultiplier && i != p.Employer {
			if producers[i].addEmployee(p, config) {
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
}

// Enum equivalent constants
const (
	FoodIdx     int = 0
	GasolineIdx int = 1
	CoffeeIdx   int = 2
)

// Adjusts the price and salary of employees based on the stock
func (p *Producer) adjustVariables() {
	newPrice := 0.0
	newSalary := 0.0
	newProduction := 0.0
	if p.Stock == 0 {
		newProduction = float64(p.MonthlyProduction) * 1.1
		newPrice = float64(p.Price) * 1.1
		newSalary = float64(p.MonthSalary) * 1.05
	} else {
		newProduction = float64(p.MonthlyProduction) * 0.9
		newPrice = float64(p.Price) * 0.9
		newSalary = float64(p.MonthSalary) * 0.95
	}
	p.MonthlyProduction = int(newProduction + 0.5)
	p.Price = int(newPrice + 0.5)
	p.MonthSalary = int(newSalary + 0.5)
}

// Adds as much product to the producer as they have money to make
func (p *Producer) produceProducts() {
	p.Stock += p.MonthlyProduction
	p.UnpaidUnits += p.MonthlyProduction
}

func (p *Producer) payProductionCost(producers []Producer, config SimConfig) {
	purchases := int(float64(p.UnpaidUnits) / float64(config.ProductionUnitCostAmount))
	p.BankBalance -= producers[GasolineIdx].registerPurchase(purchases * config.ProductionGasCost)
	p.BankBalance -= producers[CoffeeIdx].registerPurchase(purchases * config.ProductionCoffeeCost)

	p.UnpaidUnits -= purchases * config.ProductionUnitCostAmount
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

// Checks if a new employee can be hired. Employs them and returns true if so, returns false otherwise.
func (p *Producer) addEmployee(person *Person, config SimConfig) bool {
	if p.MonthHires < config.MaxHires {
		p.Employees = append(p.Employees, person)
		p.NumEmployees += 1
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
	amount := int(money / p.Price)
	if amount > p.Stock {
		return p.Stock
	}
	return amount
}

func main() {
	config, err := loadConfig()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
	}

	r := rand.New(rand.NewSource(time.Now().UnixMilli()))
	producers := make([]Producer, 3)
	producers[FoodIdx] = initProducer("food", config, r)
	producers[GasolineIdx] = initProducer("gasoline", config, r)
	producers[CoffeeIdx] = initProducer("coffee", config, r)

	people := []Person{}
	for i := 0; i < config.NumPeople; i++ {
		person := initPerson(r, i, config)

		producers[person.Employer].NumEmployees += 1
		producers[person.Employer].Employees = append(producers[person.Employer].Employees, &person)
		people = append(people, person)

	}

	detailedMonths := make([]DetailedMonth, config.MaxMonths)
	basicMonths := make([]BasicMonthTable, config.MaxMonths)
	for month := 0; month < config.MaxMonths; month++ {
		producers, people = simulationStep(producers, people, month, config)
		detailedMonth := fillDetailedMonth(people, producers, month)
		basicMonth := fillBasicMonth(people, producers, month)

		detailedMonths[month] = detailedMonth
		basicMonths[month] = basicMonth

		if month == config.MaxMonths-1 {
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
		producers[i].adjustVariables()
		producers[i].payProductionCost(producers, config)
		producers[i].produceProducts()
	}

	for i := range people {
		people[i].checkNewJobs(producers, config)
		people[i].calculateGasConsumption(producers, config)
		people[i].receiveSalary(producers)
		if month == config.PayoutMonth {
			people[i].WalletAmount *= 2
		}
		people[i].buyGoods(producers)
	}

	return producers, people
}

// Initialises a new producer and returns it
func initProducer(product string, config SimConfig, r *rand.Rand) Producer {
	stock := 0
	if config.InitWithStock {
		stock = config.InitStock
	}
	return Producer{BankBalance: config.InitBalance, Product: product, Price: config.InitPrice, Stock: stock, MonthSalary: config.InitSalary, Employees: []*Person{}, NumEmployees: 0, MonthlyProduction: config.InitMonthlyProduction, PosX: randIntInRange(config.PositionMin, config.PositionMax, r), PosY: randIntInRange(config.PositionMin, config.PositionMax, r)}
}

// Creates a new person, generating random variables. Returns the person and the producer they are employed by
func initPerson(r *rand.Rand, ID int, config SimConfig) Person {
	randomEmployer := randIntInRange(0, 3, r)

	return Person{
		IdNumber:          ID,
		Employer:          randomEmployer,
		WalletAmount:      0,
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

	exampleConfig := SimConfig{
		MaxMonths: 100, PayoutMonth: 49, NumPeople: 20, FoodIntakeMin: 30, FoodIntakeMax: 60, JobSwitchMultiplier: 1.5, InitSalary: 10, MaxHires: 2, InitBalance: 1000, InitWithStock: true, InitStock: 1000, InitPrice: 10, InitMonthlyProduction: 1000, ProductionUnitCostAmount: 10, ProductionCoffeeCost: 1, ProductionGasCost: 1, PositionMin: 0, PositionMax: 300, GasConsumptionPerDistance: 1,
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
