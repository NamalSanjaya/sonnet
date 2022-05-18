package broker

import (
	cmtype "github.com/NamalSanjayaFernando/sonnet/common/types"
)

type QueueInterface interface {
	Enqueue()
	Dequeue()
}

type Queue struct {
	Store []*cmtype.DataUnit
}

func NewQueue() *Queue{
	return &Queue{
		Store: []*cmtype.DataUnit{},
	}
}

func(q *Queue) Enqueue(elem *cmtype.DataUnit) {
	q.Store = append(q.Store,elem)
}

func(q *Queue) Dequeue()*cmtype.DataUnit {
	firstElem := q.Store[0]
	q.Store = q.Store[1:]
	return firstElem
}
