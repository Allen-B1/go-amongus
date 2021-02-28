package amongus

type StringRef Ref

func (r StringRef) Len() Int32Ref {
	return Int32Ref(Ref(r).Ref(0x8))
}

func (r StringRef) Read() string {
	size := int(r.Len().Read())
	data := make([]byte, size*2)
	Ref(r).Ref(0xC).Read(data)

	bytes := make([]byte, 0, size)
	for _, b := range data {
		if b != 0 {
			bytes = append(bytes, b)
		}
	}
	return string(bytes)
}

func (r StringRef) Write(s string) {
	size := int(r.Len().Read())
	data := make([]byte, size*2)
	for i := 0; i < size; i++ {
		data[i*2] = ' '
		data[i*2+1] = 0
	}
	i := 0
	for _, ch := range s {
		if i*2+1 >= len(data) {
			break
		}
		if len(string(ch)) == 1 {
			data[i*2] = byte(ch)
			data[i*2+1] = 0
		}
		if len(string(ch)) == 2 {
			data[i*2+1] = string(ch)[0]
			data[i*2+1] = string(ch)[1]
		}
		i++
	}
	Ref(r).Ref(0xC).Write(data)
}

type StringPtrRef PtrRef

func (r StringPtrRef) Null() bool {
	return PtrRef(r).Null()
}

func (r StringPtrRef) Deref() StringRef {
	return StringRef(PtrRef(r).Deref())
}
