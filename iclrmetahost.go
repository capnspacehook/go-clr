// +build windows

package clr

import (
	"fmt"
	"syscall"
	"unsafe"

	"golang.org/x/sys/windows"
)

// Couldnt have done any of this without this SO answer I stumbled on:
// https://stackoverflow.com/questions/37781676/how-to-use-com-component-object-model-in-golang

//ICLRMetaHost Interface from metahost.h
type ICLRMetaHost struct {
	vtbl *ICLRMetaHostVtbl
}

type ICLRMetaHostVtbl struct {
	QueryInterface                   uintptr
	AddRef                           uintptr
	Release                          uintptr
	GetRuntime                       uintptr
	GetVersionFromFile               uintptr
	EnumerateInstalledRuntimes       uintptr
	EnumerateLoadedRuntimes          uintptr
	RequestRuntimeLoadedNotification uintptr
	QueryLegacyV2RuntimeBinding      uintptr
	ExitProcess                      uintptr
}

// CLRCreateInstance provides one of three interfaces: ICLRMetaHost, ICLRMetaHostPolicy, or ICLRDebugging.
// HRESULT CLRCreateInstance(
//   [in]  REFCLSID  clsid,
//   [in]  REFIID     riid,
//   [out] LPVOID  * ppInterface
// );
// https://docs.microsoft.com/en-us/dotnet/framework/unmanaged-api/hosting/clrcreateinstance-function
func CLRCreateInstance(clsid, riid windows.GUID) (ppInterface *ICLRMetaHost, err error) {
	debugPrint("Entering into iclrmetahost.CLRCreateInstance()...")

	if clsid != CLSID_CLRMetaHost {
		err = fmt.Errorf("the input Class ID (CLSID) is not supported: %s", clsid)
		return
	}

	modMSCoree := syscall.MustLoadDLL("mscoree.dll")
	procCLRCreateInstance := modMSCoree.MustFindProc("CLRCreateInstance")

	// For some reason this procedure call returns "The specified procedure could not be found." even though it works
	hr, _, _ := procCLRCreateInstance.Call(
		uintptr(unsafe.Pointer(&clsid)),
		uintptr(unsafe.Pointer(&riid)),
		uintptr(unsafe.Pointer(&ppInterface)),
	)

	if hr != S_OK {
		err = fmt.Errorf("the mscoree!CLRCreateInstance function returned a non-zero HRESULT: 0x%x", hr)
		return
	}
	err = nil
	return
}

// GetICLRMetaHost is a wrapper function to create and return an ICLRMetahost object
func GetICLRMetaHost() (metahost *ICLRMetaHost, err error) {
	return CLRCreateInstance(CLSID_CLRMetaHost, IID_ICLRMetaHost)
}

// NewICLRMetaHost takes a uintptr to an ICLRMetahost struct in memory. This pointer should come from the syscall CLRCreateInstance
func NewICLRMetaHostFromPtr(ppv uintptr) *ICLRMetaHost {
	return (*ICLRMetaHost)(unsafe.Pointer(ppv))
}

func (obj *ICLRMetaHost) AddRef() uintptr {
	ret, _, _ := syscall.Syscall(
		obj.vtbl.AddRef,
		1,
		uintptr(unsafe.Pointer(obj)),
		0,
		0)
	return ret
}

func (obj *ICLRMetaHost) Release() uintptr {
	ret, _, _ := syscall.Syscall(
		obj.vtbl.Release,
		1,
		uintptr(unsafe.Pointer(obj)),
		0,
		0)
	return ret
}

func (obj *ICLRMetaHost) EnumerateInstalledRuntimes(pInstalledRuntimes *uintptr) uintptr {
	ret, _, _ := syscall.Syscall(
		obj.vtbl.EnumerateInstalledRuntimes,
		2,
		uintptr(unsafe.Pointer(obj)),
		uintptr(unsafe.Pointer(pInstalledRuntimes)),
		0)
	return ret
}

func (obj *ICLRMetaHost) GetRuntime(pwzVersion *uint16, riid *windows.GUID, pRuntimeHost *uintptr) uintptr {
	ret, _, _ := syscall.Syscall6(
		obj.vtbl.GetRuntime,
		4,
		uintptr(unsafe.Pointer(obj)),
		uintptr(unsafe.Pointer(pwzVersion)),
		uintptr(unsafe.Pointer(&IID_ICLRRuntimeInfo)),
		uintptr(unsafe.Pointer(pRuntimeHost)),
		0,
		0)
	return ret
}
