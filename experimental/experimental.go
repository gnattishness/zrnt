package main

import (
	"encoding/binary"
	"fmt"
	. "github.com/protolambda/zrnt/experimental/tree"
	. "github.com/protolambda/zrnt/experimental/views"
)

// Experimental code! Everything a tree and cached by default.
//  - nodes are Root (single node without children) or Commit (a node combining a pair of child nodes together)
//  - nodes are enhanced with navigation to fetch getters/setters
//  - nodes can be made read-only
//  - modifications in a subtree can be batched: do not rebind each all the way up to the root of the tree
//  - views can wrap nodes to provide typing
//  - type definitions can build new backings and wrap backings into views
//  - Vector/List/Container/Uint views supported.
//  - views can be overlaid on existing trees
//    - overlay on incomplete tree == partial
//  - Views to be implemented still:
//     - Bitvector
//     - Bitlist
//     - Union
//     - Basic-lists

type Slot uint64

func (s Slot) MerkleRoot(h HashFn) (out Root) {
	binary.LittleEndian.PutUint64(out[:], uint64(s))
	return
}

var BlockDef = &ContainerType{
	Fields: []TypeDef{
		Uint64Type,
		Uint64Type,
		BlockBodyDef,
	},
}

type Block struct {
	*ContainerView
}

func NewBlock() (b *Block) {
	return &Block{ContainerView: BlockDef.New()}
}

func (b *Block) Slot() Slot { return Slot(b.Get(0).(Uint64View)) }

var BlockBodyDef = &ContainerType{
	Fields: []TypeDef{
		Uint64Type,
		Uint64Type,
		Uint64Type,
		Uint64Type,
	},
}

type BlockBody struct {
	*ContainerView
}

func (b *Block) Body() *BlockBody {
	return &BlockBody{b.Get(2).(*ContainerView)}
}

func main() {
	b := NewBlock()
	err := b.Set(0, Uint64View(1))
	fmt.Println(err)
	fmt.Printf("%x\n", b.ViewRoot(Hash))
	err = b.Set(0, Uint64View(1))
	fmt.Println(err)
	fmt.Printf("%x\n", b.ViewRoot(Hash))
	err = b.Set(0, Uint64View(1))
	fmt.Println(err)
	fmt.Printf("%x\n", b.ViewRoot(Hash))

	fmt.Println("getting body A...")
	bodyA := b.Body()
	fmt.Printf("bodyA: %x\n", bodyA.ViewRoot(Hash))

	fmt.Println("changing body A...")
	err = bodyA.Set(0, Uint64View(1))
	fmt.Println(err)
	fmt.Printf("block: %x\n", b.ViewRoot(Hash))
	fmt.Printf("bodyA: %x\n", bodyA.ViewRoot(Hash))


	fmt.Println("getting body B...")
	bodyB := b.Body()
	fmt.Printf("bodyB: %x\n", bodyB.ViewRoot(Hash))

	fmt.Println("changing body B...")
	err = bodyB.Set(0, Uint64View(2))
	fmt.Println(err)
	fmt.Printf("block: %x\n", b.ViewRoot(Hash))
	fmt.Printf("bodyA: %x\n", bodyA.ViewRoot(Hash))
	fmt.Printf("bodyB: %x\n", bodyB.ViewRoot(Hash))

	fmt.Println("updating block with body A...")
	err = b.Set(2, bodyA)
	fmt.Println(err)
	fmt.Printf("block: %x\n", b.ViewRoot(Hash))
	fmt.Printf("bodyA: %x\n", bodyA.ViewRoot(Hash))
	fmt.Printf("bodyB: %x\n", bodyB.ViewRoot(Hash))

	fmt.Println("updating block with body B...")
	err = b.Set(2, bodyB)
	fmt.Println(err)
	fmt.Printf("block: %x\n", b.ViewRoot(Hash))
	fmt.Printf("bodyA: %x\n", bodyA.ViewRoot(Hash))
	fmt.Printf("bodyB: %x\n", bodyB.ViewRoot(Hash))
}