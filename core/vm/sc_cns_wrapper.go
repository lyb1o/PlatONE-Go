package vm

import (
	"errors"
	"fmt"

	"github.com/PlatONEnetwork/PlatONE-Go/common"
	"github.com/PlatONEnetwork/PlatONE-Go/log"
	"github.com/PlatONEnetwork/PlatONE-Go/params"
)

var (
	errInvalidCallNotFromInit = errors.New("[CNS] cnsRegisterFromInit can only be called from init()")
	errInvalidCallFromInit    = errors.New("[CNS] cnsRegister can't be called from init()")
	errLowRegVersion          = errors.New("[CNS] Version must be larger than previous version")
	errNameAndVerReg          = errors.New("[CNS] name and version is already registered and activated in CNS")
	errNameReg                = errors.New("[CNS] Name is already registered")
	errNameAndVerUnReg        = errors.New("[CNS] Name and version didn't register before")
)

type CnsWrapper struct {
	base *CnsManager
}

func (cns *CnsWrapper) RequiredGas(input []byte) uint64 {
	if common.IsBytesEmpty(input) {
		return 0
	}
	return params.CnsManagerGas
}

// Run runs the precompiled contract
func (cns *CnsWrapper) Run(input []byte) ([]byte, error) {
	defer func() {
		if e := recover(); e != nil {
			err := fmt.Errorf("[CNS] sc_cns_db.go rlp encode/decode error: %+v", e.(string))
			log.Error("[CNS] sc_cns_db.go rlp encode/decode", "error", err)
			fmt.Println(err)
		}
	}()

	return execSC(input, cns.AllExportFns())
}

// for access control
func (cns *CnsWrapper) AllExportFns() SCExportFns {
	return SCExportFns{
		"cnsRegisterFromInit":             cns.cnsRegisterFromInit,
		"cnsRegister":                     cns.cnsRegister,
		"cnsRedirect":                     cns.cnsRedirect, // cnsUnregister is deprecated, replaced by cnsRedirect
		"getContractAddress":              cns.getContractAddress,
		"ifRegisteredByAddress":           cns.ifRegisteredByAddress,
		"ifRegisteredByName":              cns.ifRegisteredByName,
		"getRegisteredContracts":          cns.getRegisteredContractsByRange,
		"getRegisteredContractsByName":    cns.getRegisteredContractsByName, // getHistoryContractsByName -> getRegisteredContractsByName
		"getRegisteredContractsByAddress": cns.getRegisteredContractsByAddress,
		"getRegisteredContractsByOrigin":  cns.getRegisteredContractsByOrigin, // getContractInfoByAddress -> getRegisteredContractsByOrigin
	}
}

func (cns *CnsWrapper) cnsRegisterFromInit(name, version string) (int32, error) {
	err := cns.base.cnsRegisterFromInit(name, version)
	return cnsRegisterErrHandle(err)
}

func (cns *CnsWrapper) cnsRegister(name, version string, address common.Address) (int32, error) {
	err := cns.base.cnsRegister(name, version, address)
	return cnsRegisterErrHandle(err)
}

func cnsRegisterErrHandle(err error) (int32, error) {

	switch err {
	case errInvalidCallFromInit, errInvalidCallNotFromInit:
		return int32(cnsInvalidCall), err
	case errNotOwner:
		return int32(cnsNoPermission), err
	case errNameInvalid, errVersionInvalid, errLowRegVersion:
		return int32(cnsInvalidArgument), err
	case errNameAndVerReg, errNameReg:
		return int32(cnsRegErr), err
	}

	return int32(cnsSuccess), nil
}

func (cns *CnsWrapper) cnsRedirect(name, version string) (int32, error) {
	err := cns.base.cnsRedirect(name, version)

	switch err {
	case errNotOwner:
		return int32(cnsNoPermission), err
	case errNameInvalid, errVersionInvalid:
		return int32(cnsInvalidArgument), err
	case errNameAndVerUnReg:
		return int32(cnsRegErr), err
	}

	return int32(cnsSuccess), nil
}

func (cns *CnsWrapper) getContractAddress(name, version string) (string, error) {
	addr, err := cns.base.getContractAddress(name, version)
	if err != nil {
		return "", err
	}

	return addr.String(), nil
}

func (cns *CnsWrapper) ifRegisteredByAddress(address common.Address) (int32, error) {
	var code int32
	isReg, _ := cns.base.ifRegisteredByAddress(address)

	if isReg {
		code = int32(cnsRegistered)
	} else {
		code = int32(cnsUnregistered)
	}

	return code, nil
}

func (cns *CnsWrapper) ifRegisteredByName(name string) (int32, error) {
	var code int32
	isReg, err := cns.base.ifRegisteredByName(name)

	if err != nil {
		return int32(cnsInvalidArgument), err
	}

	if isReg {
		code = int32(cnsRegistered)
	} else {
		code = int32(cnsUnregistered)
	}

	return code, nil
}

func (cns *CnsWrapper) getRegisteredContractsByRange(head, size int) (string, error) {
	cnsInfoArray, err := cns.base.getRegisteredContractsByRange(head, size)
	if err != nil {
		return newInternalErrorResult(err).String(), nil
	}

	return newSuccessResult(cnsInfoArray).String(), nil
}

func (cns *CnsWrapper) getRegisteredContractsByName(name string) (string, error) {
	cnsInfoArray, err := cns.base.getRegisteredContractsByName(name)
	if err != nil {
		return newInternalErrorResult(err).String(), nil
	}

	return newSuccessResult(cnsInfoArray).String(), nil
}

func (cns *CnsWrapper) getRegisteredContractsByAddress(addr common.Address) (string, error) {
	cnsInfoArray, err := cns.base.getRegisteredContractsByAddress(addr)
	if err != nil {
		return newInternalErrorResult(err).String(), nil
	}

	return newSuccessResult(cnsInfoArray).String(), nil
}

func (cns *CnsWrapper) getRegisteredContractsByOrigin(origin common.Address) (string, error) {
	cnsInfoArray, err := cns.base.getRegisteredContractsByOrigin(origin)
	if err != nil {
		return newInternalErrorResult(err).String(), nil
	}

	return newSuccessResult(cnsInfoArray).String(), nil
}