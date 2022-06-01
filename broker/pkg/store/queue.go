package store


type Interface interface {
	Enqueue()
	Dequeue()
}

type Queue struct {
	store  []*Task
	size   int
}

func NewQueue() Queue {
	return Queue{
		store: []*Task{},
		size: 0,
	}
}

func (q *Queue) Enqueue(task Task) {
	q.store = append(q.store, &task)
}

func (q *Queue) Dequeue() *Task {
	task := q.store[0]
	q.store = q.store[1:]
	return task
}

