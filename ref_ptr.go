package amongus

type PtrRef Ref

func (r PtrRef) Null() bool {
	var val uintptr
	Ref(r).Read(&val)
	return val == 0
}

func (r PtrRef) Deref() Ref {
	if r.Null() {
		panic("PtrRef.Deref called on null pointer")
	}
	return Ref(r).Ref(0, 0)
}
