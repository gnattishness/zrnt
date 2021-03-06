package proposing

import (
	"encoding/binary"
	"fmt"
	. "github.com/protolambda/zrnt/eth2/core"
	"github.com/protolambda/zrnt/eth2/meta"
	. "github.com/protolambda/zrnt/eth2/util/hashing"
	"github.com/protolambda/zrnt/eth2/util/shuffle"
)

type EpochProposerIndices [SLOTS_PER_EPOCH]ValidatorIndex

type ProposersData struct {
	Epoch   Epoch
	Current *EpochProposerIndices
	Next    *EpochProposerIndices
}

func (state *ProposersData) GetBeaconProposerIndex(slot Slot) ValidatorIndex {
	epoch := slot.ToEpoch()
	if epoch == state.Epoch {
		return state.Current[slot-state.Epoch.GetStartSlot()]
	} else if epoch == state.Epoch+1 {
		return state.Next[slot-(state.Epoch+1).GetStartSlot()]
	} else {
		panic(fmt.Errorf("slot %d not within range", slot))
	}
}

type ProposingFeature struct {
	Meta interface {
		meta.Versioning
		meta.BeaconCommittees
		meta.EffectiveBalances
		meta.ActiveIndices
		meta.CommitteeCount
		meta.EpochSeed
	}
}

func (f *ProposingFeature) LoadBeaconProposersData() (out *ProposersData) {
	currentEpoch := f.Meta.CurrentEpoch()
	return &ProposersData{
		Epoch:   currentEpoch,
		Current: f.LoadBeaconProposerIndices(currentEpoch),
		Next:    f.LoadBeaconProposerIndices(currentEpoch + 1),
	}
}

func (f *ProposingFeature) computeProposerIndex(indices []ValidatorIndex, seed Root) ValidatorIndex {
	buf := make([]byte, 32+8, 32+8)
	copy(buf[0:32], seed[:])
	for i := uint64(0); true; i++ {
		binary.LittleEndian.PutUint64(buf[32:], i)
		h := Hash(buf)
		for j := uint64(0); j < 32; j++ {
			randomByte := h[j]
			absI := ValidatorIndex(((i<<5)|j)%uint64(len(indices)))
			shuffledI := shuffle.PermuteIndex(absI, uint64(len(indices)), seed)
			candidateIndex := indices[shuffledI]
			effectiveBalance := f.Meta.EffectiveBalance(candidateIndex)
			if effectiveBalance*0xff >= MAX_EFFECTIVE_BALANCE*Gwei(randomByte) {
				return candidateIndex
			}
		}
	}
	panic("random (but balance-biased) infinite scrolling through a committee should always find a proposer")
}

// Return the beacon proposer index for the current slot
func (f *ProposingFeature) LoadBeaconProposerIndices(epoch Epoch) (out *EpochProposerIndices) {
	seedSource := f.Meta.GetSeed(epoch, DOMAIN_BEACON_PROPOSER)
	indices := f.Meta.GetActiveValidatorIndices(epoch)

	out = new(EpochProposerIndices)

	startSlot := epoch.GetStartSlot()
	for i := Slot(0); i < SLOTS_PER_EPOCH; i++ {
		buf := make([]byte, 32+8, 32+8)
		copy(buf[0:32], seedSource[:])
		binary.LittleEndian.PutUint64(buf[32:], uint64(startSlot + i))
		seed := Hash(buf)
		out[i] = f.computeProposerIndex(indices, seed)
	}
	return
}
