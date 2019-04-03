package main

import (
	"fmt"
	"math/rand"
	"time"
)

// A Task is the machine it should run on and the time taken
type Task [2]int

// An Operation is a references to a Task, e.g. {2, 1} is the first Task of Job 2
type Operation [2]int

// A Job is a list of tasks
type Job []Task

// A Shop is a list of Jobs that we will optimise for a given number of machines
type Shop struct {
	Jobs     []Job
	Machines int
}

// Data from the example given on the course page
var example = Shop{
	Jobs: []Job{
		{{0, 1}, {2, 1}, {2, 3}},
		{{0, 1}, {0, 2}, {1, 3}},
		{{1, 3}, {2, 4}},
	},
	Machines: 3,
}

// Data from the real question
var dataset = Shop{
	Jobs: []Job{
		{{0, 1}, {2, 1}, {2, 3}, {3, 3}},
		{{0, 1}, {3, 2}, {0, 2}, {1, 3}, {4, 1}},
		{{1, 3}, {2, 4}, {3, 1}, {4, 4}},
		{{2, 1}, {3, 1}, {0, 1}, {4, 1}},
	},
	Machines: 5,
}

// Gene representation of the problem is a list of jobs and a list of machines to run the job on.
// For example {1, 1, 2, 2, 3, 1, 2, 3} means run the first task of job 1, then the second task of job 1,
// then the first task of job 2, then the second task of job 2, then the first task of job 3, etc.
type Gene []int

// GeneAndFitness ties a gene to fitness
type GeneAndFitness struct {
	DNA     *Gene
	Fitness int
}

func initialisePopulation(shop Shop, populationCount int) []GeneAndFitness {
	population := []GeneAndFitness{}
	for totalPopulation := 0; totalPopulation < populationCount; totalPopulation++ {
		gene := Gene{}
		for jobIndex, job := range shop.Jobs {
			for i := 0; i < len(job); i++ {
				gene = append(gene, jobIndex+1)
			}
		}
		gene = shuffle(gene)
		fitness := totalTime(shop, &gene)
		population = append(population, GeneAndFitness{DNA: &gene, Fitness: fitness})
	}
	return population
}

func totalTime(shop Shop, gene *Gene) int {
	seen := map[int]int{}
	var schedule []Operation
	for i, x := range *gene {
		if _, found := seen[[]int(*gene)[i]]; !found {
			seen[[]int(*gene)[i]] = 1
		} else {
			seen[[]int(*gene)[i]] = seen[[]int(*gene)[i]] + 1
		}
		schedule = append(schedule, Operation{x, seen[[]int(*gene)[i]]})
	}

	var machines []machine
	for i := 0; i < shop.Machines; i++ {
		machines = append(machines, machine{})
	}
	for _, s := range schedule {
		machines[shop.Jobs[s[0]-1][s[1]-1][0]].Queue = append(machines[shop.Jobs[s[0]-1][s[1]-1][0]].Queue, s)
	}
	for time := 0; ; time++ {
		// Have we Finished?
		done := true
		// Have all machines got no time left and nothing to do?
		for i := 0; i < len(machines); i++ {
			if len(machines[i].Queue) > 0 {
				done = false
			}
			if machines[i].TimeLeft > 0 {
				done = false
			}
		}
		if done {
			return time
		}
		// Try to Queue the next task
		for i := 0; i < len(machines); i++ {
			if len(machines[i].Queue) == 0 {
				// Do nothing
			} else if machines[i].TimeLeft != 0 {
				// Do nothing
			} else if okToQueue(machines[i].Queue[0], machines) {
				// Start executing next task
				machines[i].CurrentTask, machines[i].Queue = machines[i].Queue[0], machines[i].Queue[1:]
				machines[i].TimeLeft = shop.Jobs[machines[i].CurrentTask[0]-1][machines[i].CurrentTask[1]-1][1]
			}
		}
		// Time marches onwards
		for i := 0; i < len(machines); i++ {
			if machines[i].TimeLeft > 0 {
				machines[i].TimeLeft = machines[i].TimeLeft - 1
			}
			if machines[i].TimeLeft == 0 {
				if machines[i].CurrentTask.Equals(Operation{0, 0}) {
					// Do nothing
				} else {
					machines[i].CompletedTasks = append(machines[i].CompletedTasks, machines[i].CurrentTask)
					machines[i].CurrentTask = Operation{0, 0}
				}
			}
		}
	}

}

// Crossover breeds 2 parents in a multi-point crossover fashion
func Crossover(a, b Gene) Gene {
	p1 := make(Gene, len(a))
	copy(p1, a)
	p2 := make(Gene, len(b))
	copy(p2, b)
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	length := len(p1)
	s1, s2 := r.Intn(length-2)+1, r.Intn(length-2)+1
	if s1 > s2 {
		s1, s2 = s2, s1
	}
	//println(s1, s2)
	substring := p1[s1:s2]
	//fmt.Printf("%# v\n", pretty.Formatter(substring))
	for _, x := range substring {
		for i := 0; i < len(p2); i++ {
			if p2[i] == x {
				p2 = append(p2[:i], p2[i+1:]...)
				break
			}
		}
	}
	var c Gene
	c = append(p2, substring...)
	return c
}

// Mutate has a random chance of mutating
func Mutate(g Gene) Gene {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	random := r.Intn(9)
	if random < 1 {
		i1 := r.Intn(len(g) - 1)
		i2 := r.Intn(len(g) - 1)
		g[i1], g[i2] = g[i2], g[i1]
		return g
	} else {
		return g
	}
}

func main() {
	//p := initialisePopulation(example, 2)
	//
	//totalTime(example, p[0])
	previousGeneration := initialisePopulation(dataset, 50)
	var nextGeneration []GeneAndFitness
	fmt.Printf("Generation,Mean fitness,Best fitness\n")
	for i := 1; i <= 50; i++ {
		sum := 0
		best := 10000000
		for _, x := range previousGeneration {
			sum += x.Fitness
			if x.Fitness < best {
				best = x.Fitness
			}
		}
		mean := float64(sum) / float64(len(previousGeneration))
		fmt.Printf("%d,%f,%d\n", i, mean, best)
		for i := 0; i < len(previousGeneration); i++ {
			if float64(previousGeneration[i].Fitness) <= mean-0.7 {
				if len(nextGeneration) < 50 {
					nextGeneration = append(nextGeneration, previousGeneration[i])
				}

			}
		}
		for i := 0; i < len(nextGeneration)/2; i++ {
			p1 := nextGeneration[i].DNA
			p2 := nextGeneration[i*2].DNA
			newChild := Crossover(*p1, *p2)
			newChild = Mutate(newChild)
			newMember := GeneAndFitness{DNA: &newChild, Fitness: totalTime(dataset, &newChild)}
			nextGeneration = append(nextGeneration, newMember)
		}
		previousGeneration = nextGeneration
		nextGeneration = nil
	}
	return
}

// Utility function for shuffling a gene
func shuffle(g Gene) Gene {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	for i := len(g); i > 0; i-- {
		index := r.Intn(i)
		g[i-1], g[index] = g[index], g[i-1]
	}
	return g
}

// Equals is a Utility function to see if a task is equal to another task
func (t *Operation) Equals(another Operation) bool {
	return (t[0] == another[0]) && (t[1] == another[1])
}

type machine struct {
	CompletedTasks []Operation
	CurrentTask    Operation
	Queue          []Operation
	TimeLeft       int
}

func okToQueue(o Operation, mach []machine) bool {
	if o[1] == 1 {
		return true
	}
	for i := 0; i < len(mach); i++ {
		for _, op := range mach[i].CompletedTasks {
			if o[0] == op[0] && (o[1]-1) == op[1] {
				return true
			}
		}
	}
	return false
}
