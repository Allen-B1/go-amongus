package amongus

type ListPtrRef PtrRef

func (r ListPtrRef) Null() bool {
	return PtrRef(r).Null()
}

func (r ListPtrRef) Deref() ListRef {
	return ListRef(PtrRef(r).Deref())
}

type ArrayPtrRef PtrRef

func (r ArrayPtrRef) Null() bool {
	return PtrRef(r).Null()
}

func (r ArrayPtrRef) Deref() ArrayRef {
	return ArrayRef(PtrRef(r).Deref())
}

// Type ListRef represents a reference for a List<T>
type ListRef Ref

func (r ListRef) Len() Int32Ref {
	ref := Ref(r).Ref(12)
	return Int32Ref(ref)
}

func (r ListRef) Cap() Int32Ref {
	ref := Ref(r).Ref(8, 12)
	return Int32Ref(ref)
}

// Method Items returns a reference to the first element of the array.
func (r ListRef) Items() Ref {
	ref := Ref(r).Ref(8, 16)
	return ref
}

type ArrayRef Ref

func (r ArrayRef) Len() Int32Ref {
	ref := Ref(r).Ref(12)
	return Int32Ref(ref)
}

func (r ArrayRef) Items() Ref {
	ref := Ref(r).Ref(16)
	return ref
}
