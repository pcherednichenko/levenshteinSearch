package levenshtein

// D is the levenshtein distance calculator interface
type D interface {
	// Dist calculates levenshtein distance between two utf-8 encoded strings
	Dist(string, string) int
}

// New creates a new levenshtein distance calculator where indel is increment/deletion cost
// and sub is the substitution cost.
func New(indel, sub int) D {
	return &calculator{indel, sub}
}

type calculator struct {
	indel, sub int
}

// https://en.wikibooks.org/wiki/Algorithm_Implementation/Strings/Levenshtein_distance#C
func (c *calculator) Dist(s1, s2 string) int {
	l := len(s1)
	m := make([]int, l+1)
	for i := 1; i <= l; i++ {
		m[i] = i * c.indel
	}
	lastdiag, x, y := 0, 1, 1
	var a, b, d int
	for _, rx := range s2 {
		m[0], lastdiag, y = x*c.indel, (x-1)*c.indel, 1
		for _, ry := range s1 {
			a = m[y] + c.indel
			b = m[y-1] + c.indel
			d = lastdiag + c.subCost(&rx, &ry)
			m[y], lastdiag = min3(&a, &b, &d), m[y]
			y++
		}
		x++
	}
	return m[l]
}

func (c *calculator) subCost(r1 *int32, r2 *rune) int {
	if *r1 == *r2 {
		return 0
	}
	return c.sub
}

func min3(a *int, b *int, c *int) int {
	d := min(b, c)
	return min(a, &d)
}

func min(a *int, b *int) int {
	if *a < *b {
		return *a
	}
	return *b
}

var defaultCalculator = New(1, 1)

// Dist is a convenience function for a levenshtein distance calculator with equal costs.
func Dist(s1, s2 string) int {
	return defaultCalculator.Dist(s1, s2)
}
