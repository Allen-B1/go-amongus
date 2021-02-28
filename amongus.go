package amongus

//go:generate go run generate/generate_ref_basic.go
//go:generate go run generate/generate_ref_class.go v2020.12.9s

import (
	"bytes"
	"fmt"
	"reflect"
	"unsafe"

	"github.com/mitchellh/go-ps"
	windows "golang.org/x/sys/windows"
)

var (
	ErrNotOpen = fmt.Errorf("Among Us not open")
)

var (
	kernel32               = windows.MustLoadDLL("kernel32.dll")
	procReadProcessMemory  = kernel32.MustFindProc("ReadProcessMemory")
	procWriteProcessMemory = kernel32.MustFindProc("WriteProcessMemory")
	procModule32First      = kernel32.MustFindProc("Module32First")
	procModule32Next       = kernel32.MustFindProc("Module32Next")
	procCloseHandle        = kernel32.MustFindProc("CloseHandle")
)

type AmongUs struct {
	handle      windows.Handle
	ModBaseAddr uintptr
	ModBaseSize uintptr
}

type Ref struct {
	AU   *AmongUs
	Addr uintptr
}

func (r Ref) Read(data interface{}) {
	ptr := unsafe.Pointer(reflect.ValueOf(data).Pointer())
	size := reflect.TypeOf(data).Elem().Size()
	if reflect.TypeOf(data).Kind() == reflect.Array || reflect.TypeOf(data).Kind() == reflect.Slice {
		size *= uintptr(reflect.ValueOf(data).Cap())
	}
	ret, _, err := procReadProcessMemory.Call(uintptr(r.AU.handle), r.Addr, uintptr(ptr), size, 0)
	if ret == 0 {
		panic(err)
	}
}

func (r Ref) Write(data interface{}) {
	ptr := unsafe.Pointer(reflect.ValueOf(data).Pointer())
	size := reflect.TypeOf(data).Elem().Size()
	if reflect.TypeOf(data).Kind() == reflect.Array || reflect.TypeOf(data).Kind() == reflect.Slice {
		size *= uintptr(reflect.ValueOf(data).Cap())
	}
	ret, _, err := procWriteProcessMemory.Call(uintptr(r.AU.handle), r.Addr, uintptr(ptr), size, 0)
	if ret == 0 {
		panic(err)
	}
}

func (r Ref) Ref(offsets ...uintptr) Ref {
	ref, err := r.TryRef(offsets...)
	if err != nil {
		panic(err)
	}
	return ref
}

func (r Ref) TryRef(offsets ...uintptr) (Ref, error) {
	var addr = r.Addr & 0xffffffff

	if len(offsets) == 0 {
		return r, nil
	}

	for _, offset := range offsets[:len(offsets)-1] {
		ret, _, err := procReadProcessMemory.Call(uintptr(r.AU.handle), addr+offset, uintptr(unsafe.Pointer(&addr)), unsafe.Sizeof(addr), 0)
		if ret == 0 {
			return Ref{}, err
		}
		addr = addr & 0xffffffff
	}

	return Ref{r.AU, addr + offsets[len(offsets)-1]}, nil
}

// Function Ref creates a reference.
//
// Reference identification algorithm:
//
//     Read at of base + offsets[0]
//     Read at previous value + offsets[1]
//     Read at previous value + offsets[2]
//     ...
//     Read at previous value + offsets[n-1]
//     Return previous value + offsets[n]
func (au *AmongUs) Ref(baseAddress uintptr, offsets ...uintptr) Ref {
	ref, err := au.TryRef(baseAddress, offsets...)
	if err != nil {
		panic(err)
	}
	return ref
}

func (au *AmongUs) TryRef(baseAddress uintptr, offsets ...uintptr) (Ref, error) {
	ref := Ref{au, baseAddress & 0xffffffff}
	if len(offsets) != 0 {
		var err error
		ref, err = ref.TryRef(offsets...)
		if err != nil {
			return Ref{}, err
		}
	}

	return ref, nil
}

func (au *AmongUs) Find(memory []byte, from uintptr, to uintptr, found func(Ref) bool) (Ref, bool) {
	blockAddr := uintptr(from)
	var buf [4096]byte
	for {
		ret, _, err := procReadProcessMemory.Call(uintptr(au.handle), blockAddr, uintptr(unsafe.Pointer(&buf[0])), uintptr(len(buf)), 0)
		if ret != 0 {
			if i := bytes.Index(buf[:], memory); i != -1 {
				addr := blockAddr + uintptr(i)
				ref := Ref{au, addr}
				if found(ref) {
					return ref, true
				}
			}
		} else {
			_ = err
		}

		blockAddr += 4096

		if blockAddr > to {
			break
		}
	}

	return Ref{}, false
}

type structMODULEENTRY32 struct {
	DwSize        uint32
	Th32ModuleID  uint32
	Th32ProcessID uint32
	GlblcntUsage  uint32
	ProccntUsage  uint32
	ModBaseAddr   uintptr
	ModBaseSize   uint32
	HMODULE       uintptr
	SzModule      [256]byte
	SzExePath     [260]byte
}

func New() (*AmongUs, error) {
	pslist, err := ps.Processes()
	if err != nil {
		panic(err)
	}
	auPid := int32(-1)
	for _, proc := range pslist {
		if proc.Executable() == "Among Us.exe" {
			auPid = int32(proc.Pid())
			break
		}
	}
	if auPid < 0 {
		return nil, ErrNotOpen
	}
	return NewFromPID(uint32(auPid))
}

func NewFromPID(pid uint32) (*AmongUs, error) {
	handle, err := windows.OpenProcess(0x0008|0x0010|0x0020, false, pid)
	if err != nil {
		return nil, fmt.Errorf("unable to get Among Us process: %w", err)
	}

	snapHandle, err := windows.CreateToolhelp32Snapshot(windows.TH32CS_SNAPMODULE32|windows.TH32CS_SNAPMODULE, pid)
	if err != nil {
		return nil, fmt.Errorf("unable to create snapshot: %w", err)
	}

	var data structMODULEENTRY32
	data.DwSize = uint32(unsafe.Sizeof(data))
	ret, _, err := procModule32First.Call(uintptr(snapHandle), uintptr(unsafe.Pointer(&data)))
	if ret == 0 {
		return nil, fmt.Errorf("unable to create snapshot: %w", err)
	}

	au := new(AmongUs)
	au.handle = handle

	for {
		i := bytes.IndexByte(data.SzModule[:], 0)
		mod := string(data.SzModule[:i])

		if mod == "GameAssembly.dll" {
			au.ModBaseAddr = data.ModBaseAddr
			au.ModBaseSize = uintptr(data.ModBaseSize)
		}

		ret, _, _ := procModule32Next.Call(uintptr(snapHandle), uintptr(unsafe.Pointer(&data)))
		if ret == 0 {
			break
		}
	}

	return au, nil
}
