package main

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
)

type Node struct {
	Value       int
	Left, Right *Node
}

func NewNode(val int) (result *Node) {
	result = new(Node)
	result.Value = val
	return
}

func (target *Node) InsertRight(node *Node) {
	if node.Left != nil || node.Right != nil {
		panic("Inserting attached node")
	}
	if target == nil {
		node.Left = node
		node.Right = node
	} else {
		node.Right = target.Right
		node.Left = target
		target.Right.Left = node
		target.Right = node
	}
	return
}

func (target *Node) InsertLeft(node *Node) {
	if target == nil {
		target.InsertRight(node)
	}
	target.Left.InsertRight(node)
}

func (node *Node) String() (result string) {
	first := true
	var prev *Node
	for start := node; node != nil; node = node.Right {
		if !first && node == start {
			return result + "..."
		}

		if !first && node.Left != prev {
			panic("Inconsistency detected")
		}

		result += fmt.Sprintf("%d, ", node.Value)

		prev = node
		first = false
	}
	panic("List is not circular")
}

func (node *Node) ShiftLeft(shift int) *Node {
	for i := 0; i < shift; i++ {
		node = node.Left
	}
	return node
}

func (node *Node) ShiftRight(shift int) *Node {
	for i := 0; i < shift; i++ {
		node = node.Right
	}
	return node
}

func (node *Node) Unattach() {
	l, r := node.Left, node.Right
	l.Right = r
	r.Left = l
	node.Left = nil
	node.Right = nil
}

func (node *Node) Move(shift int) {
	if shift == 0 || node == node.Left || node == node.Right {
		return
	}

	if shift > 0 {
		target := node.Right
		node.Unattach()
		shift--
		target = target.ShiftRight(shift)
		target.InsertRight(node)
	} else {
		shift = -shift
		target := node.Left
		node.Unattach()
		shift--
		target = target.ShiftLeft(shift)
		target.InsertLeft(node)
	}
	return
}

func ParseInput(scanner *bufio.Scanner) (node0 *Node, nodes []*Node) {
	var end *Node
	nodes = []*Node{}

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		num, err := strconv.Atoi(line)
		if err != nil {
			panic(err)
		}

		node := NewNode(num)
		end.InsertRight(node)
		nodes = append(nodes, node)
		end = node

		if num == 0 {
			node0 = node
		}
	}

	if end == nil {
		panic("Empty list")
	}

	if node0 == nil {
		panic("Node with coordinate 0 not found")
	}

	return
}

func main() {
	scanner := bufio.NewScanner(os.Stdin)

	node0, nodes := ParseInput(scanner)

	// fmt.Printf("Initial: %s\n", nodes[0])
	for i := 0; i < len(nodes); i++ {
		shift := nodes[i].Value
		nodes[i].Move(shift)

		// fmt.Printf("%d (shift %d): %s\n", i, shift, nodes[0])
	}

	n1 := node0.ShiftRight(1000)
	n2 := n1.ShiftRight(1000)
	n3 := n2.ShiftRight(1000)
	fmt.Println(n1.Value + n2.Value + n3.Value)
}
