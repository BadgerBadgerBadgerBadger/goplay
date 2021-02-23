package main

import "math/rand"

const numRageEmojis = 5

var rageEmojis = [numRageEmojis]string{"😤", "😡", "😠", "🤬", "👿"}

// Rager provides angry emojis. An instance or rager, on subsequent calls of the
// Rand method, yields up a new random angry emoji.
type Rager struct {
	length int
	unseen []int
}

func NewRager() Rager {

	length := numRageEmojis

	rager := Rager{
		length: length,
	}
	rager.initUnseen()

	return rager
}

// initUnseen initializes a new slice of indices into rageEmojis,
// and then shuffles them.
func (r *Rager) initUnseen() {

	unseen := make([]int, r.length)

	for i := 0; i < r.length; i++ {
		unseen[i] = i
	}

	rand.Shuffle(r.length, func(i, j int) {
		unseen[i], unseen[j] = unseen[j], unseen[i]
	})

	r.unseen = unseen
}

// Rand returns a new angry emoji from a predetermined pool of
// angry emojis. It tries to provide the best chance that two
// of the same emoji won't be yielded repeatedly.
func (r *Rager) Rand() string {

	// let's grab the very first emoji
	index := r.unseen[0]

	if len(r.unseen) == 1 {
		// if the list only has 1 index, we've consumed it
		// so let's refresh the indices
		r.initUnseen()
	} else {
		// otherwise we drop the first index
		r.unseen = r.unseen[1:]
	}

	// and we return an emoji
	return rageEmojis[index]
}
