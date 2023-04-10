package common

type (
	CallFunc func(args CallArgs) (res interface{}, err error)

	CallArgs []interface{}

	Functor struct {
		CallFunc  CallFunc
		CallArgs  CallArgs
		NoRelease bool
	}

	DecorateHook func(functor *Functor) (res interface{}, err error)
)

func NewDecorateFunc(functor *Functor, hook DecorateHook) *Functor {
	df := &Functor{
		CallFunc:  decorateFunc,
		NoRelease: functor.NoRelease,
	}
	df.AddArgs(functor)
	df.AddArgs(hook)
	return df
}

func decorateFunc(args CallArgs) (res interface{}, err error) {
	functor := args[0].(*Functor)
	hook := args[1].(DecorateHook)
	add_cnt := 0
	for idx, arg := range args {
		if idx > 1 {
			functor.AddArgs(arg)
			add_cnt += 1
		}
	}
	res, err = hook(functor)
	if functor.NoRelease {
		for i := 0; i < add_cnt; i++ {
			functor.PopArgs()
		}
	}
	return
}

func (f *Functor) Call() (res interface{}, err error) {
	res, err = f.CallFunc(f.CallArgs)
	if !f.NoRelease {
		f.Release()
	}
	return
}

func (f *Functor) CallWithAddArgs(args ...interface{}) (res interface{}, err error) {
	for _, addArg := range args {
		f.CallArgs = append(f.CallArgs, addArg)
	}
	return f.Call()
}

func (f *Functor) AddArgs(args ...interface{}) {
	for _, addArg := range args {
		f.CallArgs = append(f.CallArgs, addArg)
	}
}

func (f *Functor) PopArgs() {
	f.CallArgs = f.CallArgs[:len(f.CallArgs)-1]
}

func (f *Functor) Release() {
	f.CallArgs = nil
	f.CallFunc = nil
}
