package notify

type Notifier interface {
	notify()
}

func Notify(n Notifier) {
	n.notify()
}
