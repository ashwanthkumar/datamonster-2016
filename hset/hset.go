package hset

// MapSetBasedHeap - Map based Set implementation
// also implements container/heap.Interface for keeping the largest T occuring keys
type MapSetBasedHeap struct {
	_data map[string]int
	keys  []string
}

// Empty set of strings
func Empty() MapSetBasedHeap {
	return MapSetBasedHeap{
		_data: make(map[string]int),
		keys:  []string{},
	}
}

// FromSlice - Creates a new Set from a slice of strings
func FromSlice(slice []string) MapSetBasedHeap {
	set := Empty()
	for _, elem := range slice {
		set.Add(elem)
	}

	return set
}

// Contains - Checks for an existence of an element
func (s MapSetBasedHeap) Contains(elem string) bool {
	_, present := s._data[elem]
	return present
}

// Add an element to the Set
func (s *MapSetBasedHeap) Add(elem string) {
	i, present := s._data[elem]
	if !present {
		s.keys = append(s.keys, elem)
		i = 0
	}
	i++
	s._data[elem] = i
}

// Union another Set to this set and returns that
func (s MapSetBasedHeap) Union(another MapSetBasedHeap) MapSetBasedHeap {
	union := FromSlice(s.Values())
	for _, value := range another.Values() {
		union.Add(value)
	}
	return union
}

// Intersect another Set to this Set and returns that
func (s MapSetBasedHeap) Intersect(another MapSetBasedHeap) MapSetBasedHeap {
	intersection := Empty()
	for _, elem := range another.Values() {
		if s.Contains(elem) {
			intersection.Add(elem)
		}
	}
	return intersection
}

// IsSupersetOf another Set
func (s MapSetBasedHeap) IsSupersetOf(another MapSetBasedHeap) bool {
	found := true
	for _, elem := range another.Values() {
		found = found && s.Contains(elem)
	}
	return found
}

// Equal another Set
func (s MapSetBasedHeap) Equal(another MapSetBasedHeap) bool {
	found := s.Size() == another.Size()
	if found {
		for _, elem := range another.Values() {
			found = found && s.Contains(elem)
		}
	}
	return found
}

// Remove an element if it exists in the Set
// Returns if the value was present and removed
func (s *MapSetBasedHeap) Remove(elem string) bool {
	found := s.Contains(elem)
	delete(s._data, elem)
	if found {
		var i = 0
		for index := 0; index < s.Size(); index++ {
			if s.keys[index] == elem {
				i = index
				break
			}
		}
		s.keys = append(s.keys[:i], s.keys[i+1:]...)
	}
	return found
}

// Values of the underlying set
func (s MapSetBasedHeap) Values() []string {
	var values []string
	for key := range s._data {
		values = append(values, key)
	}

	return values
}

// Size of the set
func (s MapSetBasedHeap) Size() int {
	return len(s._data)
}

// Len - Length of the Set
func (s MapSetBasedHeap) Len() int { return s.Size() }

// Less - sort.Interface
func (s MapSetBasedHeap) Less(i, j int) bool {
	// fmt.Printf("%v\n", s.keys)
	// fmt.Printf("%v\n", s._data)
	// fmt.Printf("i=%d,j=%d\n", i, j)
	return s._data[s.keys[i]] < s._data[s.keys[j]]
}

// Swap - sort.Interface
func (s *MapSetBasedHeap) Swap(i, j int) {
	s.keys[i], s.keys[j] = s.keys[j], s.keys[i]
}

// Push an element into the Heap
func (s *MapSetBasedHeap) Push(x interface{}) {
	s.Add(x.(string))
}

// Pop the lowest occuring item out of the Heap
func (s *MapSetBasedHeap) Pop() interface{} {
	n := s.Len()
	elem := s.keys[n-1]
	s.Remove(elem)
	return elem
}

// MaxOccuringItem - Return the most occuring item
func (s MapSetBasedHeap) MaxOccuringItem() string {
	n := s.Len()
	return s.keys[n-1]
}
