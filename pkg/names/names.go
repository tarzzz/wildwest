package names

import (
	"fmt"
	"math/rand"
	"strings"
	"sync"
	"time"
)

// NameGenerator generates unique persona names
type NameGenerator struct {
	used  map[string]bool
	mutex sync.Mutex
	rng   *rand.Rand
}

// NewNameGenerator creates a new name generator
func NewNameGenerator() *NameGenerator {
	return &NameGenerator{
		used: make(map[string]bool),
		rng:  rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

// Famous names categorized by field
var (
	// Scientists
	Scientists = []string{
		"einstein", "curie", "tesla", "newton", "darwin",
		"hawking", "feynman", "bohr", "galileo", "kepler",
		"faraday", "maxwell", "planck", "heisenberg", "schrodinger",
		"turing", "lovelace", "franklin", "nobel", "pasteur",
	}

	// Artists
	Artists = []string{
		"picasso", "monet", "davinci", "michelangelo", "rembrandt",
		"vangogh", "dali", "kahlo", "warhol", "banksy",
		"basquiat", "klimt", "matisse", "kandinsky", "pollock",
		"rothko", "hopper", "okeeffe", "cassatt", "renoir",
	}

	// Musicians
	Musicians = []string{
		"mozart", "beethoven", "bach", "chopin", "vivaldi",
		"tchaikovsky", "brahms", "debussy", "handel", "wagner",
		"hendrix", "lennon", "presley", "armstrong", "coltrane",
		"davis", "ellington", "parker", "mingus", "monk",
	}

	// Writers
	Writers = []string{
		"shakespeare", "hemingway", "tolkien", "austen", "dickens",
		"twain", "wilde", "orwell", "kafka", "poe",
		"woolf", "joyce", "nabokov", "dostoyevsky", "tolstoy",
		"marquez", "borges", "cervantes", "homer", "dante",
	}

	// Philosophers
	Philosophers = []string{
		"socrates", "plato", "aristotle", "kant", "nietzsche",
		"descartes", "locke", "hume", "spinoza", "wittgenstein",
		"russell", "sartre", "camus", "kierkegaard", "epicurus",
		"aquinas", "confucius", "laozi", "buddha", "seneca",
	}

	// Inventors & Engineers
	Inventors = []string{
		"edison", "bell", "wright", "marconi", "diesel",
		"gutenberg", "watt", "ford", "nobel", "archimedes",
		"brunel", "hopper", "lovelace", "babbage", "jacquard",
		"stephenson", "musk", "jobs", "gates", "berners-lee",
	}

	// Explorers & Pioneers
	Explorers = []string{
		"magellan", "columbus", "armstrong", "shackleton", "hillary",
		"cousteau", "gagarin", "earhart", "polo", "livingstone",
	}

	// Quality & Testing Experts
	QualityExperts = []string{
		"deming", "shewhart", "juran", "crosby", "ishikawa",
		"taguchi", "feigenbaum", "ohno", "imai", "hopper",
	}

	// All names combined
	AllNames = []string{}
)

func init() {
	// Combine all name lists
	AllNames = append(AllNames, Scientists...)
	AllNames = append(AllNames, Artists...)
	AllNames = append(AllNames, Musicians...)
	AllNames = append(AllNames, Writers...)
	AllNames = append(AllNames, Philosophers...)
	AllNames = append(AllNames, Inventors...)
	AllNames = append(AllNames, Explorers...)
	AllNames = append(AllNames, QualityExperts...)
}

// GetRandomName returns a random unused name
func (ng *NameGenerator) GetRandomName() string {
	ng.mutex.Lock()
	defer ng.mutex.Unlock()

	// Get available names
	available := ng.getAvailableNames()
	if len(available) == 0 {
		// If all names used, append timestamp
		return fmt.Sprintf("persona-%d", time.Now().Unix())
	}

	// Pick random available name
	name := available[ng.rng.Intn(len(available))]
	ng.used[name] = true
	return name
}

// GetNameByCategory returns a random name from a specific category
func (ng *NameGenerator) GetNameByCategory(category string) string {
	ng.mutex.Lock()
	defer ng.mutex.Unlock()

	var pool []string
	category = strings.ToLower(category)

	switch category {
	case "scientist", "scientists":
		pool = Scientists
	case "artist", "artists":
		pool = Artists
	case "musician", "musicians":
		pool = Musicians
	case "writer", "writers":
		pool = Writers
	case "philosopher", "philosophers":
		pool = Philosophers
	case "inventor", "inventors", "engineer", "engineers":
		pool = Inventors
	case "explorer", "explorers":
		pool = Explorers
	case "qa", "quality", "tester", "testers":
		pool = QualityExperts
	default:
		pool = AllNames
	}

	// Filter available names from pool
	available := []string{}
	for _, name := range pool {
		if !ng.used[name] {
			available = append(available, name)
		}
	}

	if len(available) == 0 {
		// Fallback to any available name
		return ng.GetRandomName()
	}

	name := available[ng.rng.Intn(len(available))]
	ng.used[name] = true
	return name
}

// GetNameForPersona returns an appropriate name based on persona type
func (ng *NameGenerator) GetNameForPersona(personaType string) string {
	personaType = strings.ToLower(personaType)

	if strings.Contains(personaType, "manager") {
		// Managers get philosopher or leader names
		return ng.GetNameByCategory("philosopher")
	} else if strings.Contains(personaType, "architect") {
		// Architects get artist or inventor names
		if ng.rng.Intn(2) == 0 {
			return ng.GetNameByCategory("artist")
		}
		return ng.GetNameByCategory("inventor")
	} else if strings.Contains(personaType, "engineer") {
		// Engineers get scientist or inventor names
		if ng.rng.Intn(2) == 0 {
			return ng.GetNameByCategory("scientist")
		}
		return ng.GetNameByCategory("inventor")
	} else if strings.Contains(personaType, "intern") {
		// Interns get writer or explorer names
		if ng.rng.Intn(2) == 0 {
			return ng.GetNameByCategory("writer")
		}
		return ng.GetNameByCategory("explorer")
	} else if strings.Contains(personaType, "qa") {
		// QA gets quality expert names
		return ng.GetNameByCategory("qa")
	}

	return ng.GetRandomName()
}

// getAvailableNames returns list of unused names
func (ng *NameGenerator) getAvailableNames() []string {
	available := []string{}
	for _, name := range AllNames {
		if !ng.used[name] {
			available = append(available, name)
		}
	}
	return available
}

// MarkUsed marks a name as used
func (ng *NameGenerator) MarkUsed(name string) {
	ng.mutex.Lock()
	defer ng.mutex.Unlock()
	ng.used[strings.ToLower(name)] = true
}

// IsAvailable checks if a name is available
func (ng *NameGenerator) IsAvailable(name string) bool {
	ng.mutex.Lock()
	defer ng.mutex.Unlock()
	return !ng.used[strings.ToLower(name)]
}

// GetNameList returns all available names by category
func GetNameList() map[string][]string {
	return map[string][]string{
		"Scientists":     Scientists,
		"Artists":        Artists,
		"Musicians":      Musicians,
		"Writers":        Writers,
		"Philosophers":   Philosophers,
		"Inventors":      Inventors,
		"Explorers":      Explorers,
		"QualityExperts": QualityExperts,
	}
}

// CountTotal returns total number of names
func CountTotal() int {
	return len(AllNames)
}

// GetRandomNameStatic returns a random name without tracking (for one-off use)
func GetRandomNameStatic() string {
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))
	return AllNames[rng.Intn(len(AllNames))]
}
