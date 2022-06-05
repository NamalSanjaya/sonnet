package store


type Interface interface {
	Enqueue(task string)
	Dequeue() string
	Len() int
	IsEmpty() bool
}

type Queue struct {
	store  []string
}

var _ Interface = (*Queue)(nil)

func NewQueue() *Queue {
	return &Queue{
		store: []string{},
	}
}

func (q *Queue) Enqueue(task string) {
	q.store = append(q.store, task)
}

func (q *Queue) Dequeue() string {
	task := q.store[0]
	q.store = q.store[1:]
	return task
}

func (q *Queue) Len() int {
	return len(q.store)
}

func (q *Queue) IsEmpty() bool {
	return len(q.store) == 0 
}
