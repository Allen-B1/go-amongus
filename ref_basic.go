package amongus
type Int8Ref Ref

func (r Int8Ref) Read() int8 {
	var i int8
	Ref(r).Read(&i)
	return i
}

func (r Int8Ref) Write(i int8) {
	Ref(r).Write(&i)
}

type Uint8Ref Ref

func (r Uint8Ref) Read() uint8 {
	var i uint8
	Ref(r).Read(&i)
	return i
}

func (r Uint8Ref) Write(i uint8) {
	Ref(r).Write(&i)
}

type Int16Ref Ref

func (r Int16Ref) Read() int16 {
	var i int16
	Ref(r).Read(&i)
	return i
}

func (r Int16Ref) Write(i int16) {
	Ref(r).Write(&i)
}

type Uint16Ref Ref

func (r Uint16Ref) Read() uint16 {
	var i uint16
	Ref(r).Read(&i)
	return i
}

func (r Uint16Ref) Write(i uint16) {
	Ref(r).Write(&i)
}

type Int32Ref Ref

func (r Int32Ref) Read() int32 {
	var i int32
	Ref(r).Read(&i)
	return i
}

func (r Int32Ref) Write(i int32) {
	Ref(r).Write(&i)
}

type Uint32Ref Ref

func (r Uint32Ref) Read() uint32 {
	var i uint32
	Ref(r).Read(&i)
	return i
}

func (r Uint32Ref) Write(i uint32) {
	Ref(r).Write(&i)
}

type Int64Ref Ref

func (r Int64Ref) Read() int64 {
	var i int64
	Ref(r).Read(&i)
	return i
}

func (r Int64Ref) Write(i int64) {
	Ref(r).Write(&i)
}

type Uint64Ref Ref

func (r Uint64Ref) Read() uint64 {
	var i uint64
	Ref(r).Read(&i)
	return i
}

func (r Uint64Ref) Write(i uint64) {
	Ref(r).Write(&i)
}

type Float32Ref Ref

func (r Float32Ref) Read() float32 {
	var i float32
	Ref(r).Read(&i)
	return i
}

func (r Float32Ref) Write(i float32) {
	Ref(r).Write(&i)
}

type Float64Ref Ref

func (r Float64Ref) Read() float64 {
	var i float64
	Ref(r).Read(&i)
	return i
}

func (r Float64Ref) Write(i float64) {
	Ref(r).Write(&i)
}

type BoolRef Ref

func (r BoolRef) Read() bool {
	var i bool
	Ref(r).Read(&i)
	return i
}

func (r BoolRef) Write(i bool) {
	Ref(r).Write(&i)
}
