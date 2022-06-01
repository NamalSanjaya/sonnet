package store


type Interface interface {
	Enqueue(task Task)
	Dequeue() *Task
	Len() int
	IsEmpty() bool
}

type Queue struct {
	store  []*Task
}

var _ Interface = (*Queue)(nil)

func NewQueue() *Queue {
	return &Queue{
		store: []*Task{},
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

func (q *Queue) Len() int {
	return len(q.store)
}

func (q *Queue) IsEmpty() bool {
	return len(q.store) == 0 
}
