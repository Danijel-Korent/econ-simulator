package main

import (
	"fmt"
	"math/rand"
)

type Person struct {
	IdNumber          int
	Employer          int
	WalletAmount      int
	Salary            int
	MonthlyFoodIntake int
	MonthlyGasIntake  int
	PosX              int
	PosY              int

	FoodConsumption   int
	GasConsumption    int
	CoffeeConsumption int
	SavingsRatio      float64
}

// Creates a new person, generating random variables. Returns the person and the producer they are employed by
func initPerson(r *rand.Rand, ID int, config SimConfig) Person {
	randomEmployer := randIntInRange(0, len(config.Producers), r)

	return Person{
		IdNumber:          ID,
		Employer:          randomEmployer,
		WalletAmount:      randIntInRange(config.StartingWalletMin, config.StartingWalletMax, r),
		SavingsRatio:      randFloatInRange(config.SavingsRatioMin, config.SavingsRatioMax, r),
		Salary:            0,
		MonthlyFoodIntake: randIntInRange(config.FoodIntakeMin, config.FoodIntakeMax, r),
		PosX:              randIntInRange(config.PositionMin, config.PositionMax, r),
		PosY:              randIntInRange(config.PositionMin, config.PositionMax, r),
	}
}

func (p *Person) simulationStep(producers []Producer, config SimConfig) {
	p.receiveSalary(producers)
	p.calculateGasConsumption(producers, config)
	p.buyGoods(producers)
}

func (p *Person) setWalletAmount(amount int) {
	if amount < 0 {
		panic(fmt.Sprintf("Attempted to set wallet amount of person %v to %v", p.IdNumber, amount))
	}

	p.WalletAmount = amount
}

// Simulates the purchase of goods, adjusting variables on the person and alerting the producer
func (p *Person) buyGoods(producers []Producer) {
	//Savings (temporarily subtracted and then re-added after purchases)
	savings := int(float64(p.WalletAmount) * p.SavingsRatio)
	p.setWalletAmount(p.WalletAmount - savings)

	foodProducerIdx := findProducerIdx("food", producers)
	gasProducerIdx := findProducerIdx("gasoline", producers)
	coffeeProducerIdx := findProducerIdx("coffee", producers)

	p.FoodConsumption = p.buyGood(&producers[foodProducerIdx], p.MonthlyFoodIntake)
	p.GasConsumption = p.buyGood(&producers[gasProducerIdx], p.MonthlyFoodIntake)

	if p.WalletAmount > p.MonthlyFoodIntake*producers[foodProducerIdx].Price {
		p.FoodConsumption += p.buyGood(&producers[foodProducerIdx], p.MonthlyFoodIntake)
	}

	maxCoffee := producers[coffeeProducerIdx].getMaxUnits(p.WalletAmount)
	coffeeCost := producers[coffeeProducerIdx].registerPurchase(maxCoffee)
	p.setWalletAmount(p.WalletAmount - coffeeCost)
	p.CoffeeConsumption = maxCoffee

	p.setWalletAmount(p.WalletAmount + savings)
}

// Purchases a good from a producer - according to the desired intake or maximum possible
func (p *Person) buyGood(producer *Producer, intake int) int {
	foodUnits := p.getUnitsToPurchase(*producer, intake)
	foodCost := producer.registerPurchase(foodUnits)
	p.setWalletAmount(p.WalletAmount - foodCost)
	return foodUnits

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

func (p *Person) receiveSalary(producers []Producer) {
	p.Salary = producers[p.Employer].MonthSalary
	p.setWalletAmount(p.WalletAmount + p.Salary)
	producers[p.Employer].setBankBalance(producers[p.Employer].BankBalance - p.Salary)
}

func (p *Person) calculateGasConsumption(producers []Producer, config SimConfig) {
	employer := producers[p.Employer]
	p.MonthlyGasIntake = pythagDistance(p.PosX, p.PosY, employer.PosX, employer.PosY) * config.GasConsumptionPerDistance
}
