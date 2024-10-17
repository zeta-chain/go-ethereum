package vm

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/params"
)

type StatefulPrecompiledContract interface {
	ContractRef
	// RequiredPrice calculates the contract gas used
	RequiredGas(input []byte) uint64
	// Run runs the precompiled contract
	Run(evm *EVM, contract *Contract, readonly bool) ([]byte, error)
}

// Compile-time assertion to ensure StatefulPrecompiledContractWrapper implements PrecompiledContract
var _ PrecompiledContract = (*statefulPrecompiledContractWrapper)(nil)

type statefulPrecompiledContractWrapper struct {
	evm      *EVM
	contract StatefulPrecompiledContract
	readOnly bool
	caller   ContractRef
}

func (w *statefulPrecompiledContractWrapper) RequiredGas(input []byte) uint64 {
	return w.contract.RequiredGas(input)
}

func (w *statefulPrecompiledContractWrapper) Run(input []byte) ([]byte, error) {
	c := &Contract{
		CallerAddress: w.caller.Address(),
		Input:         input,
		caller:        w.caller,
		self:          w.contract,
		isPrecompile:  true,
	}
	return w.contract.Run(w.evm, c, w.readOnly)
}

// SetPrecompiles sets the precompiled contracts for the EVM.
// It is not thread-safe.
func (evm *EVM) SetStatefulPrecompiles(precompiles []StatefulPrecompiledContract) {
	evm.statefulPrecompiles = map[common.Address]StatefulPrecompiledContract{}
	for _, p := range precompiles {
		evm.statefulPrecompiles[p.Address()] = p
	}
}

// AllPrecompiledAddresses gets both builtin precompiles and the extra stateful precompiles
func (evm *EVM) AllPrecompiledAddresses(rules params.Rules) []common.Address {
	allPrecompiles := ActivePrecompiles(rules)
	for addr := range evm.statefulPrecompiles {
		allPrecompiles = append(allPrecompiles, addr)
	}
	return allPrecompiles
}
