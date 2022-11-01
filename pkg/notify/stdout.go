package notify

type StdoutNotifier struct{}

func NewStdoutNotifier() StdoutNotifier {
	return StdoutNotifier{}
}

func (s StdoutNotifier) notify() {}
